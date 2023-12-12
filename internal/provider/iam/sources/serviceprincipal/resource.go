// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serviceprincipal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/validators"
)

func NewResource() resource.Resource {
	return &resourceConfig{}
}

type resourceConfig struct {
	client *clients.Client
}

var _ resource.Resource = &resourceConfig{}

func (r *resourceConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_principal"
}

func (r *resourceConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
The service principal resource manages an HCP Service Principal.

The user or service account that is running Terraform when creating an ` + "`hcp_service_principal` resource must have `roles/Admin`" + ` on the parent resource; either the project or organization.`,
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal's unique identitier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal's resource name in the format `iam/project/<project_id>/service-principal/<name>` or `iam/organization/<organization_id>/service-principal/<name>`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The service principal's name.",
				Validators: []validator.String{
					validators.ResourceNamePart(),
					stringvalidator.LengthBetween(3, 36),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parent": schema.StringAttribute{
				Description: `
The parent location to create the service principal under.
If unspecified, the service principal will be created in the project the provider is configured with.
If specified, the accepted values are ` + "`project/<project_id>` or `organization/<organization_id>`",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(organization|project)/.+$`),
						"must reference a project or organization resource_name",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceConfig) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type resourceModel struct {
	ResourceID   types.String `tfsdk:"resource_id"`
	ResourceName types.String `tfsdk:"resource_name"`
	Name         types.String `tfsdk:"name"`
	Parent       types.String `tfsdk:"parent"`
}

func (r *resourceConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parent := plan.Parent.ValueString()
	if parent == "" {
		parent = fmt.Sprintf("project/%s", r.client.Config.ProjectID)
	}

	createParams := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalParams()
	createParams.ParentResourceName = parent
	createParams.Body = service_principals_service.ServicePrincipalsServiceCreateServicePrincipalBody{
		Name: plan.Name.ValueString(),
	}

	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceCreateServicePrincipal(createParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service principal", err.Error())
		return
	}

	// Get parent from created SP
	sp := res.GetPayload().ServicePrincipal
	parent, err = spParent(sp.ResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving service principal parent", err.Error())
	}

	plan.ResourceID = types.StringValue(sp.ID)
	plan.ResourceName = types.StringValue(sp.ResourceName)
	plan.Name = types.StringValue(sp.Name)
	plan.Parent = types.StringValue(parent)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

// spParent extracts the parent resource name from the service principal name
func spParent(resourceName string) (string, error) {
	err := fmt.Errorf("unexpected format for service principal resource_name: %q", resourceName)
	parts := strings.SplitN(resourceName, "/", 5)
	if len(parts) != 5 || parts[0] != "iam" ||
		(parts[1] != "project" && parts[1] != "organization") ||
		parts[3] != "service-principal" {
		return "", err
	}

	return strings.Join(parts[1:3], "/"), nil
}

func (r *resourceConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParams()
	getParams.ResourceName = state.ResourceName.ValueString()
	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceGetServicePrincipal(getParams, nil)
	if err != nil {
		var getErr *service_principals_service.ServicePrincipalsServiceGetServicePrincipalDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error retrieving service principal", err.Error())
		return
	}

	// Get parent from created SP
	sp := res.GetPayload().ServicePrincipal
	parent, err := spParent(sp.ResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving service principal parent", err.Error())
	}

	state.ResourceID = types.StringValue(sp.ID)
	state.ResourceName = types.StringValue(sp.ResourceName)
	state.Name = types.StringValue(sp.Name)
	state.Parent = types.StringValue(parent)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported
}

func (r *resourceConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteParams := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalParams()
	deleteParams.ResourceName = state.ResourceName.ValueString()
	_, err := r.client.ServicePrincipals.ServicePrincipalsServiceDeleteServicePrincipal(deleteParams, nil)
	if err != nil {
		var getErr *service_principals_service.ServicePrincipalsServiceDeleteServicePrincipalDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting service principal", err.Error())
		return
	}
}

func (r *resourceConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_name"), req, resp)
}
