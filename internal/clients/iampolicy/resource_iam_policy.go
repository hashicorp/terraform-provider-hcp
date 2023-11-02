package iampolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	basePolicySchema = map[string]schema.Attribute{
		"policy_data": schema.StringAttribute{
			CustomType:  PolicyDataType{},
			Required:    true,
			Description: "The policy to apply.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"etag": schema.StringAttribute{
			Computed:    true,
			Description: "The etag captures the existing state of the policy.",
		},
	}
)

// NewResourceIamPolicy creates a new Terraform Resource for definitively
// managing the IAM Policy for the given resource. By implementing
// NewResourceIamUpdaterFunc, the resource will inherit all functionality needed
// to allow Terraform to manage the IAM Policy for the resource.
//
// The typeName is the type that supports IAM ("project", "organization",
// "vault_secrets_app", etc).

// parentSpecificSchema should be a schema that includes a MarkdownDescription
// and any necessary Attributes to target the specific resource ("project_id",
// "resource_name", etc)
//
// importAttrName allows specifying the attribute to be set when a user runs
// `terraform import`. Subsequent calls to SetResourceIamPolicy can use this
// information to populate the policy.
func NewResourceIamPolicy(
	typeName string,
	parentSpecificSchema schema.Schema,
	importAttrName string,
	newUpdaterFunc NewResourceIamUpdaterFunc,
) resource.Resource {
	return &resourcePolicy{
		parentSchema:   parentSpecificSchema,
		typeName:       typeName,
		importAttrName: importAttrName,
		updaterFunc:    newUpdaterFunc,
	}
}

type resourcePolicy struct {
	parentSchema   schema.Schema
	typeName       string
	importAttrName string
	updaterFunc    NewResourceIamUpdaterFunc
	client         *clients.Client
}

func (r *resourcePolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_iam_policy", req.ProviderTypeName, r.typeName)
}

func (r *resourcePolicy) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourcePolicy) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.parentSchema.MarkdownDescription == "" {
		resp.Diagnostics.AddError("Parent Schema did not include a MarkdownDescription", "Please report this issue to the provider developers.")
	}

	// Validate the parent schema doesn't implement anything that is in the
	// base schema
	for k := range basePolicySchema {
		if _, ok := r.parentSchema.Attributes[k]; ok {
			resp.Diagnostics.AddError("Parent Schema attributes overlap with base schema", "Please report this issue to the provider developers.")
		}
	}
}

func (r *resourcePolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.parentSchema.MarkdownDescription,
		Attributes:          r.parentSchema.Attributes,
	}

	if resp.Schema.Attributes == nil || len(resp.Schema.Attributes) == 0 {
		resp.Schema.Attributes = basePolicySchema
	} else {
		for k, v := range basePolicySchema {
			resp.Schema.Attributes[k] = v
		}
	}
}

func (r *resourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	updater, diags := r.updaterFunc(ctx, &req.Plan, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	// Copy the existing state and then override with the new policy. This
	// allows the attributes set by the ResourceIamUpdater to carry forward.
	resp.State.Raw = req.Plan.Raw
	resp.Diagnostics.Append(setIamPolicyData(ctx, &req.Plan, &resp.State, updater)...)
}

func (r *resourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	updater, diags := r.updaterFunc(ctx, &req.State, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	p, diags := updater.GetResourceIamPolicy(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(storeIamPolicyData(ctx, &resp.State, p)...)
}

func (r *resourcePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updater, diags := r.updaterFunc(ctx, &req.Plan, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	// Copy the existing state and then override with the new policy. This
	// allows the attributes set by the ResourceIamUpdater to carry forward.
	resp.State.Raw = req.Plan.Raw
	resp.Diagnostics.Append(setIamPolicyData(ctx, &req.Plan, &resp.State, updater)...)
}

func (r *resourcePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	updater, diags := r.updaterFunc(ctx, &req.State, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	// Delete by setting an empty policy
	p := &models.HashicorpCloudResourcemanagerPolicy{}
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("etag"), &p.Etag)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newPolicy, diags := updater.SetResourceIamPolicy(ctx, p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(storeIamPolicyData(ctx, &resp.State, newPolicy)...)
}

func (r *resourcePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.importAttrName != "" {
		resource.ImportStatePassthroughID(ctx, path.Root(r.importAttrName), req, resp)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("etag"), "")...)
	}
}

func setIamPolicyData(ctx context.Context, in, out TerraformResourceData, updater ResourceIamUpdater) diag.Diagnostics {

	// Get the encoded policy
	var diags diag.Diagnostics
	var encodedPolicyData PolicyDataValue
	diags.Append(in.GetAttribute(ctx, path.Root("policy_data"), &encodedPolicyData)...)
	if diags.HasError() {
		return diags
	}

	// Unmarshall it
	var p models.HashicorpCloudResourcemanagerPolicy
	if err := p.UnmarshalBinary([]byte(encodedPolicyData.ValueString())); err != nil {
		diags.AddError("failed to unmarshal policy_data", err.Error())
		return diags
	}

	// If the etag is not set, we need to fetch and set it
	if p.Etag == "" {
		existingPolicy, getDiags := updater.GetResourceIamPolicy(ctx)
		diags.Append(getDiags...)
		if diags.HasError() {
			return diags
		}

		p.Etag = existingPolicy.Etag
	}

	updatedPolicy, setDiags := updater.SetResourceIamPolicy(ctx, &p)
	diags.Append(setDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(storeIamPolicyData(ctx, out, updatedPolicy)...)
	return diags
}

func storeIamPolicyData(ctx context.Context, d TerraformResourceData, p *models.HashicorpCloudResourcemanagerPolicy) diag.Diagnostics {
	// Marshal the policy
	var diags diag.Diagnostics

	// Extract the etag and clear it
	etag := p.Etag
	p.Etag = ""

	raw, err := p.MarshalBinary()
	if err != nil {
		diags.AddError("failed to set updated policy_data",
			fmt.Sprintf("Please report this issue to the provider developers: %s", err.Error()))
		return diags
	}

	// Store the updated policy
	d.SetAttribute(ctx, path.Root("etag"), etag)
	d.SetAttribute(ctx, path.Root("policy_data"), PolicyDataValue{
		StringValue: basetypes.NewStringValue(string(raw)),
	})
	return diags
}
