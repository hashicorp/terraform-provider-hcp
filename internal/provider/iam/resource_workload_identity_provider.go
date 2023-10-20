package iam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	sso "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

func NewWorkloadIdentityProviderResource() resource.Resource {
	return &resourceWorkloadIdentityProvider{}
}

type resourceWorkloadIdentityProvider struct {
	client *clients.Client
}

func (r *resourceWorkloadIdentityProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_workload_identity_provider"
}

func (r *resourceWorkloadIdentityProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The workload identity provider resource allows federating an external identity to a HCP Service Principal.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The workload identity provider's name. Ideally, this should be descriptive of the workload being federated.",
				Validators: []validator.String{
					hcpvalidator.ResourceNamePart(),
					stringvalidator.LengthBetween(3, 36),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_principal": schema.StringAttribute{
				Description: "The service principal's resource name for which the workload identity provider will be created for. Only service principals created within a project are allowed.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^iam/project/.+/service-principal/.+$`),
						"must reference a project service principal resource_name.",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description for the workload identity provider.",
				Validators: []validator.String{
					hcpvalidator.ResourceNamePart(),
					stringvalidator.LengthBetween(0, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"conditional_access": schema.StringAttribute{
				Required: true,
				Description: "conditional_access is a hashicorp/go-bexpr string " +
					"that is evaluated when exchanging tokens. It restricts which upstream " +
					"identities are allowed to access the service principal.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(5, 511),
				},
			},
			"aws": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Required:    true,
						Description: "The AWS Account ID that is allowed to exchange workload identities.",
						Validators: []validator.String{
							stringvalidator.LengthBetween(12, 12),
						},
					},
				},
				Optional: true,
				Validators: []validator.Object{
					// Validate only this attribute or oidc is configured.
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("oidc"),
					}...),
				},
			},
			"oidc": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"issuer_uri": schema.StringAttribute{
						Required:    true,
						Description: "The URL of the OIDC Issuer that is allowed to exchange workload identities.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^https://.*\..+/$`),
								"must be a valid URL starting with https:// and must end in /",
							),
						},
					},
					"allowed_audiences": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Description: "allowed_audiences is the set of audiences set on the access " +
							"token that are allowed to exchange identities. The access token must have an " +
							"audience that is contained in this set. If no audience is set, the default " +
							"allowed audience will be the resource name of the WorkloadIdentityProvider.",
						Default: setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.LengthBetween(1, 511),
							),
						},
					},
				},
				Optional: true,
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The workload identity providers's unique identitier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("The workload identity providers's resource name in the format `%s`",
					"iam/project/<project_id>/service-principal/<sp_name>/workload-identity-provider/<name>"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceWorkloadIdentityProvider) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}

type WorkloadIdentityProvider struct {
	Name              types.String `tfsdk:"name"`
	ServicePrincipal  types.String `tfsdk:"service_principal"`
	Description       types.String `tfsdk:"description"`
	ConditionalAccess types.String `tfsdk:"conditional_access"`
	AWS               types.Object `tfsdk:"aws"`
	OIDC              types.Object `tfsdk:"oidc"`
	ResourceID        types.String `tfsdk:"resource_id"`
	ResourceName      types.String `tfsdk:"resource_name"`

	aws  *AWSProvider  `tfsdk:"-"`
	oidc *OIDCProvider `tfsdk:"-"`
}

type AWSProvider struct {
	AccountID types.String `tfsdk:"account_id"`
}

func (a AWSProvider) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_id": types.StringType,
	}
}

type OIDCProvider struct {
	IssuerURL        types.String `tfsdk:"issuer_uri"`
	AllowedAudiences types.Set    `tfsdk:"allowed_audiences"`
	allowedAudiences []string     `tfsdk:"-"`
}

func (o OIDCProvider) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"issuer_uri":        types.StringType,
		"allowed_audiences": types.SetType{ElemType: types.StringType},
	}
}

// extract extracts the Go values form their Terraform wrapped values.
func (w *WorkloadIdentityProvider) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !w.AWS.IsNull() {
		w.aws = &AWSProvider{}
		diags = w.AWS.As(ctx, w.aws, basetypes.ObjectAsOptions{})
	} else if !w.OIDC.IsNull() {
		w.oidc = &OIDCProvider{}
		diags = w.OIDC.As(ctx, w.oidc, basetypes.ObjectAsOptions{})

		w.oidc.allowedAudiences = make([]string, 0, len(w.oidc.AllowedAudiences.Elements()))
		diags2 := w.oidc.AllowedAudiences.ElementsAs(ctx, &w.oidc.allowedAudiences, false)
		diags = append(diags, diags2...)
	}

	return diags
}

// fromModel encodes the values from a WorkloadIdentityProvider model into the
// Terraform values, such that they can be saved to state.
func (w *WorkloadIdentityProvider) fromModel(ctx context.Context, wip *models.HashicorpCloudIamWorkloadIdentityProvider) diag.Diagnostics {
	var diags diag.Diagnostics

	parts := strings.SplitN(wip.ResourceName, "/", 8)
	if len(parts) != 7 {
		diags.AddError("unexpected workload identity provider resource name", wip.ResourceName)
		return diags
	}

	w.Name = types.StringValue(parts[6])
	w.ServicePrincipal = types.StringValue(strings.Join(parts[:5], "/"))
	w.Description = types.StringValue(wip.Description)
	w.ConditionalAccess = types.StringValue(wip.ConditionalAccess)
	w.ResourceID = types.StringValue(wip.ResourceID)
	w.ResourceName = types.StringValue(wip.ResourceName)

	if wip.AwsConfig != nil {
		w.AWS = types.ObjectValueMust(w.AWS.AttributeTypes(ctx), map[string]attr.Value{
			"account_id": types.StringValue(wip.AwsConfig.AccountID),
		})
	}

	if wip.OidcConfig != nil {
		allowedAudiences, d := types.SetValueFrom(ctx, types.StringType, wip.OidcConfig.AllowedAudiences)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		w.OIDC = types.ObjectValueMust(w.OIDC.AttributeTypes(ctx), map[string]attr.Value{
			"issuer_uri":        types.StringValue(wip.OidcConfig.IssuerURI),
			"allowed_audiences": allowedAudiences,
		})
	}

	return diags
}

func (r *resourceWorkloadIdentityProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkloadIdentityProvider
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createParams := sso.NewServicePrincipalsServiceCreateWorkloadIdentityProviderParams()
	createParams.ParentResourceName = plan.ServicePrincipal.ValueString()
	createParams.Body = sso.ServicePrincipalsServiceCreateWorkloadIdentityProviderBody{
		Name: plan.Name.ValueString(),
		Provider: &models.HashicorpCloudIamWorkloadIdentityProvider{
			ConditionalAccess: plan.ConditionalAccess.ValueString(),
			Description:       plan.Description.ValueString(),
		},
	}

	if plan.aws != nil {
		createParams.Body.Provider.AwsConfig = &models.HashicorpCloudIamAWSWorkloadIdentityProviderConfig{
			AccountID: plan.aws.AccountID.ValueString(),
		}
	} else {
		createParams.Body.Provider.OidcConfig = &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
			AllowedAudiences: plan.oidc.allowedAudiences,
			IssuerURI:        plan.oidc.IssuerURL.ValueString(),
		}
	}

	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceCreateWorkloadIdentityProvider(createParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workload identity provider", err.Error())
		return
	}

	wip := res.GetPayload().Provider
	resp.Diagnostics.Append(plan.fromModel(ctx, wip)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWorkloadIdentityProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkloadIdentityProvider
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := sso.NewServicePrincipalsServiceGetWorkloadIdentityProviderParams()
	getParams.ResourceName2 = state.ResourceName.ValueString()
	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceGetWorkloadIdentityProvider(getParams, nil)
	if err != nil {
		var getErr *sso.ServicePrincipalsServiceGetWorkloadIdentityProviderDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error retrieving workload identity provider", err.Error())
		return
	}

	wip := res.GetPayload().Provider
	resp.Diagnostics.Append(state.fromModel(ctx, wip)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWorkloadIdentityProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkloadIdentityProvider
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)

	updateParams := sso.NewServicePrincipalsServiceUpdateWorkloadIdentityProviderParams()
	updateParams.ProviderResourceName = plan.ResourceName.ValueString()
	updateParams.Provider = sso.ServicePrincipalsServiceUpdateWorkloadIdentityProviderBody{
		ConditionalAccess: plan.ConditionalAccess.ValueString(),
		Description:       plan.Description.ValueString(),
	}

	if plan.aws != nil {
		updateParams.Provider.AwsConfig = &models.HashicorpCloudIamAWSWorkloadIdentityProviderConfig{
			AccountID: plan.aws.AccountID.ValueString(),
		}
	} else {
		updateParams.Provider.OidcConfig = &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
			AllowedAudiences: plan.oidc.allowedAudiences,
			IssuerURI:        plan.oidc.IssuerURL.ValueString(),
		}
	}

	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceUpdateWorkloadIdentityProvider(updateParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workload identity provider", err.Error())
		return
	}

	wip := res.GetPayload().Provider
	resp.Diagnostics.Append(plan.fromModel(ctx, wip)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWorkloadIdentityProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkloadIdentityProvider
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteParams := sso.NewServicePrincipalsServiceDeleteWorkloadIdentityProviderParams()
	deleteParams.ResourceName4 = state.ResourceName.ValueString()
	_, err := r.client.ServicePrincipals.ServicePrincipalsServiceDeleteWorkloadIdentityProvider(deleteParams, nil)
	if err != nil {
		var getErr *sso.ServicePrincipalsServiceDeleteWorkloadIdentityProviderDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting workload identity provider", err.Error())
		return
	}
}

func (r *resourceWorkloadIdentityProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_name"), req, resp)
}
