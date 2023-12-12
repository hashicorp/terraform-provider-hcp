// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serviceprincipalkey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewResource() resource.Resource {
	return &resourceConfig{}
}

type resourceConfig struct {
	client *clients.Client
}

var _ resource.Resource = &resourceConfig{}

func (r *resourceConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_principal_key"
}

func (r *resourceConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `The service principal key resource manages an HCP Service Principal key.

The user or service account that is running Terraform when creating an ` + "`hcp_service_principal_key` resource must have `roles/Admin`" + ` on the parent resource; either the project or organization.`,
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal key's resource name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The generated service principal client_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The generated service principal client_secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_principal": schema.StringAttribute{
				Required:    true,
				Description: "The service principal's resource name for which a key should be created.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^iam/(organization|project)/.+/service-principal/.+$`),
						"must reference a service principal resource_name",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rotate_triggers": schema.MapAttribute{
				Optional: true,
				Description: "A map of arbitrary string key/value pairs that will force recreation " +
					"of the key when they change, enabling key based on external conditions such " +
					"as a rotating timestamp. Changing this forces a new resource to be created.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
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
	ResourceName     types.String `tfsdk:"resource_name"`
	ClientID         types.String `tfsdk:"client_id"`
	ClientSecret     types.String `tfsdk:"client_secret"`
	ServicePrincipal types.String `tfsdk:"service_principal"`
	RotateTriggers   types.Map    `tfsdk:"rotate_triggers"`
}

func (r *resourceConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createParams := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalKeyParams()
	createParams.ParentResourceName = plan.ServicePrincipal.ValueString()
	res, err := r.client.ServicePrincipals.ServicePrincipalsServiceCreateServicePrincipalKey(createParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service principal key", err.Error())
		return
	}

	// Get parent from created SP
	plan.ResourceName = types.StringValue(res.Payload.Key.ResourceName)
	plan.ClientID = types.StringValue(res.Payload.Key.ClientID)
	plan.ClientSecret = types.StringValue(res.Payload.ClientSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParams()
	getParams.ResourceName = state.ServicePrincipal.ValueString()
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

	found := false
	for _, spk := range res.Payload.Keys {
		if spk.ResourceName == state.ResourceName.ValueString() {
			found = true
			break
		}
	}

	// The Service Principal no longer contains the key
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
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

	deleteParams := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalKeyParams()
	deleteParams.ResourceName2 = state.ResourceName.ValueString()
	_, err := r.client.ServicePrincipals.ServicePrincipalsServiceDeleteServicePrincipalKey(deleteParams, nil)
	if err != nil {
		var getErr *service_principals_service.ServicePrincipalsServiceDeleteServicePrincipalDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting service principal key", err.Error())
		return
	}
}
