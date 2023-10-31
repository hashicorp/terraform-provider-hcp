package iampolicy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	baseBindingSchema = map[string]schema.Attribute{
		"principal_id": schema.StringAttribute{
			Required:    true,
			Description: "The principal to bind to the given role.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role": schema.StringAttribute{
			Required:    true,
			Description: "The role name to bind to the given principal.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^roles/.+$`),
					"must reference a role name.",
				),
			},
		},
	}
)

// NewResourceIamBinding creates a new Terraform Resource for managing IAM
// Bindings for the given resource. By implementing NewResourceIamUpdaterFunc,
// the resource will inherit all functionality needed to allow IAM Bindings.
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
func NewResourceIamBinding(
	typeName string,
	parentSpecificSchema schema.Schema,
	importAttrName string,
	newUpdaterFunc NewResourceIamUpdaterFunc,
) resource.Resource {
	return &resourceBinding{
		parentSchema:   parentSpecificSchema,
		typeName:       typeName,
		importAttrName: importAttrName,
		updaterFunc:    newUpdaterFunc,
	}
}

type resourceBinding struct {
	parentSchema   schema.Schema
	typeName       string
	importAttrName string
	updaterFunc    NewResourceIamUpdaterFunc
	client         *clients.Client
}

func (r *resourceBinding) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_iam_binding", req.ProviderTypeName, r.typeName)
}

func (r *resourceBinding) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceBinding) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.parentSchema.MarkdownDescription,
		Attributes:          r.parentSchema.Attributes,
	}

	if resp.Schema.Attributes == nil || len(resp.Schema.Attributes) == 0 {
		resp.Schema.Attributes = baseBindingSchema
	} else {
		for k, v := range baseBindingSchema {
			resp.Schema.Attributes[k] = v
		}
	}
}

func getBinding(ctx context.Context, d TerraformResourceData) (*models.HashicorpCloudResourcemanagerPolicyBinding, diag.Diagnostics) {
	var p, role types.String
	diags := d.GetAttribute(ctx, path.Root("principal_id"), &p)
	diags.Append(d.GetAttribute(ctx, path.Root("role"), &role)...)
	if diags.HasError() {
		return nil, diags
	}

	return &models.HashicorpCloudResourcemanagerPolicyBinding{
		Members: []*models.HashicorpCloudResourcemanagerPolicyBindingMember{
			{
				MemberID:   p.ValueString(),
				MemberType: nil, // Will be populated in a batch look
			},
		},
		RoleID: role.ValueString(),
	}, diags
}

func (r *resourceBinding) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	updater, diags := r.updaterFunc(ctx, &req.Plan, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	binding, diags := getBinding(ctx, &req.Plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Modify the policy using the batcher and wait on the future
	_, diags = bindingsBatcher.
		getBatch(updater).
		ModifyPolicy(ctx, r.client, binding, nil).
		Get()

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Copy the existing state.
	resp.State.Raw = req.Plan.Raw
}

func (r *resourceBinding) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	updater, diags := r.updaterFunc(ctx, &req.State, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	// Get the policy using the batcher and wait on the future
	ep, diags := bindingsBatcher.
		getBatch(updater).
		GetPolicy(ctx).
		Get()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Check if the binding is in the policy
	bindings := ToMap(ep)
	binding, diags := getBinding(ctx, &req.State)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	found := false
	if members, ok := bindings[binding.RoleID]; ok {
		if _, ok := members[binding.Members[0].MemberID]; ok {
			found = true
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
	}
}

func (r *resourceBinding) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updater, diags := r.updaterFunc(ctx, &req.Plan, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	// Remove the existing binding
	rmBinding, diags := getBinding(ctx, &req.State)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Add the new binding
	updateBinding, diags := getBinding(ctx, &req.Plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Modify the policy using the batcher and wait on the future
	_, diags = bindingsBatcher.
		getBatch(updater).
		ModifyPolicy(ctx, r.client, updateBinding, rmBinding).
		Get()

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Copy the existing state.
	resp.State.Raw = req.Plan.Raw
}

func (r *resourceBinding) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	updater, diags := r.updaterFunc(ctx, &req.State, r.client)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"failed to initiate policy updater",
			"Please report this issue to the provider developers")
		return
	}

	binding, diags := getBinding(ctx, &req.State)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Modify the policy using the batcher and wait on the future
	_, diags = bindingsBatcher.
		getBatch(updater).
		ModifyPolicy(ctx, r.client, nil, binding).
		Get()

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}
