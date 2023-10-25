package resourcemanager

import (
	"context"
	"fmt"
	"regexp"

	iam "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	iamModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"golang.org/x/exp/maps"
)

const (
	// maxIAMPrincipalBindings captures the maximum number of principals that
	// can be bound in a given policy.
	maxIAMPrincipalBindings = 1000
)

type DataSourceIAMPolicy struct {
	client *clients.Client
}

type DataSourceIAMPolicyModel struct {
	Bindings   types.Set    `tfsdk:"bindings"`
	PolicyData types.String `tfsdk:"policy_data"`
	bindings   []*Binding
}

func (d *DataSourceIAMPolicyModel) extract(ctx context.Context) diag.Diagnostics {
	d.bindings = make([]*Binding, 0, len(d.Bindings.Elements()))
	return d.Bindings.ElementsAs(ctx, &d.bindings, false)
}

type Binding struct {
	Role       types.String   `tfsdk:"role"`
	Principals []types.String `tfsdk:"principals"`
}

func NewIAMPolicyDataSource() datasource.DataSource {
	return &DataSourceIAMPolicy{}
}

func (d *DataSourceIAMPolicy) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (d *DataSourceIAMPolicy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates an IAM policy that may be referenced by and applied to other HCP IAM resources, such as the `hcp_project_iam_policy` resource.",
		Attributes: map[string]schema.Attribute{
			"bindings": schema.SetNestedAttribute{
				Description: "A binding associates a set of principals to a role.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							Required:    true,
							Description: "The role name to bind to the given principals.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^roles/.+$`),
									"must reference a role name.",
								),
							},
						},
						"principals": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "The set of principals to bind to the given role.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
			"policy_data": schema.StringAttribute{
				Description: "The policy data in a format suitable for reference by resources that support setting IAM policy.",
				Computed:    true,
			},
		},
	}
}

func (d *DataSourceIAMPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client

}

func (d *DataSourceIAMPolicy) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data DataSourceIAMPolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(data.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles := make(map[string]struct{}, len(data.bindings))
	var principals int
	for i, b := range data.bindings {
		// Determine the number of unique principals being bound
		principals += len(b.Principals)

		if _, ok := roles[b.Role.ValueString()]; ok {
			p := path.Root("bindings").AtSetValue(data.Bindings.Elements()[i])
			resp.Diagnostics.AddAttributeError(p, "Duplicate role definition", fmt.Sprintf("binding for role %s already defined", b.Role))
		}

		roles[b.Role.ValueString()] = struct{}{}
	}

	if principals > maxIAMPrincipalBindings {
		resp.Diagnostics.AddError("Too many principals bound in the policy",
			fmt.Sprintf("A maximum of %d principals may be bound", maxIAMPrincipalBindings))
	}

	return

}

func (d *DataSourceIAMPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceIAMPolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(data.extract(ctx)...)

	// Gather all the principals
	principalSet := make(map[string]*iamModels.HashicorpCloudIamPrincipal, 256)
	for _, b := range data.bindings {
		for _, p := range b.Principals {
			principalSet[p.ValueString()] = nil
		}
	}

	// We don't need to paginate since the max number of principals allowed in a
	// policy is <= the allowed batch size
	params := iam.NewIamServiceBatchGetPrincipalsParams()
	params.OrganizationID = d.client.Config.OrganizationID
	params.View = (*string)(iamModels.HashicorpCloudIamPrincipalViewPRINCIPALVIEWBASIC.Pointer())
	params.PrincipalIds = maps.Keys(principalSet)

	res, err := d.client.IAM.IamServiceBatchGetPrincipals(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up principals in policy", err.Error())
		return
	}

	// Map the principal ID to the principal object
	for _, p := range res.Payload.Principals {
		principalSet[p.ID] = p
	}

	// Build the policy object
	var policy models.HashicorpCloudResourcemanagerPolicy
	for _, binding := range data.bindings {
		b := &models.HashicorpCloudResourcemanagerPolicyBinding{
			RoleID:  binding.Role.ValueString(),
			Members: make([]*models.HashicorpCloudResourcemanagerPolicyBindingMember, len(binding.Principals)),
		}

		for i, p := range binding.Principals {
			principal, ok := principalSet[p.ValueString()]
			if !ok {
				resp.Diagnostics.AddError(
					"Failed to determine principal information in IAM Policy Binding",
					"Please report this issue to the provider developers.",
				)
			}

			m := &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID: principal.ID,
			}

			switch *principal.Type {
			case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUSER:
				m.MemberType = models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer()
			case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEGROUP:
				m.MemberType = models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeGROUP.Pointer()
			case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE:
				m.MemberType = models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeSERVICEPRINCIPAL.Pointer()
			default:
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unsupported principal type (%s) for IAM Policy", *principal.Type),
					"Please report this issue to the provider developers.",
				)
			}

			b.Members[i] = m
		}

		policy.Bindings = append(policy.Bindings, b)
	}

	// Serialize the policy
	policyJSON, err := policy.MarshalBinary()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to serialize IAM Policy",
			fmt.Sprintf("Please report this issue to the provider developers. Error: %v", err),
		)
	}

	data.PolicyData = types.StringValue(string(policyJSON))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
