// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"

	service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/secret_manager_service"
)

var (
	_ resource.Resource              = &secretManagerResource{}
	_ resource.ResourceWithConfigure = &secretManagerResource{}
)

var (
	baseSecretManagerSchema = map[string]schema.Attribute{
		"project_id": schema.StringAttribute{
			Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
)

// secretManagerResource is an implementation for configuring specific types Radar secret managers.
// Examples: hcp_resource_radar_secret_manager_vault_dedicated make use of
// this implementation to define resources with specific schemas, validation, and state details related to their types.
type secretManagerResource struct {
	client                    *clients.Client
	TypeName                  string
	SecretManagerType         string
	ResourceSchema            schema.Schema
	GetSecretManagerFromPlan  func(ctx context.Context, plan tfsdk.Plan) (secretManager, diag.Diagnostics)
	GetSecretManagerFromState func(ctx context.Context, state tfsdk.State) (secretManager, diag.Diagnostics)
}

func (r *secretManagerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.TypeName
}

func (r *secretManagerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.ResourceSchema.MarkdownDescription,
		Attributes:          r.ResourceSchema.Attributes,
	}

	for k, v := range baseSecretManagerSchema {
		// check to see if schema key already exists; if so, skip
		if _, exists := resp.Schema.Attributes[k]; exists {
			continue
		}
		resp.Schema.Attributes[k] = v
	}
}

// secretManager is the minimal plan/state that a Radar secret manager must have.
type secretManager interface {
	GetProjectID() types.String
	SetProjectID(types.String)
	GetID() types.String
	SetID(types.String)
	GetConnectionURL() types.String
	GetToken() types.String
	GetAuthMethod() types.String
	SetFeatures(map[string]interface{})
	GetFeatures() map[string]interface{}
}

// base abstraction of Radar secret manager, partially implements secretManager interface
type abstractSecretManagerModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	ID        types.String `tfsdk:"id"`
}

func (b *abstractSecretManagerModel) GetProjectID() types.String { return b.ProjectID }

func (b *abstractSecretManagerModel) SetProjectID(projectID types.String) { b.ProjectID = projectID }

func (b *abstractSecretManagerModel) GetID() types.String { return b.ID }

func (b *abstractSecretManagerModel) SetID(id types.String) { b.ID = id }

func (r *secretManagerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *secretManagerResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *secretManagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	src, diags := r.GetSecretManagerFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	connection := src.GetConnectionURL().ValueString()

	var name string
	if i := strings.Index(src.GetConnectionURL().ValueString(), ":"); i >= 0 {
		name = connection[:i]
	} else {
		name = connection
	}

	var authMethod string
	if src.GetAuthMethod().IsNull() || src.GetAuthMethod().IsUnknown() {
		// This should be caught by schema validation, but just in case.
		resp.Diagnostics.AddError("Error creating Radar secret manager", "auth_method must be specified")
		return
	}
	authMethod = src.GetAuthMethod().ValueString()

	var token string
	if src.GetToken().IsNull() || src.GetToken().IsUnknown() {
		// This should be caught by schema validation, but just in case.
		resp.Diagnostics.AddError("Error creating Radar secret manager", "auth_method details must be specified")
	}
	token = src.GetToken().ValueString()

	body := service.OnboardSecretManagerBody{
		DetectorType:  "agent",
		Type:          r.SecretManagerType,
		Name:          name,
		ConnectionURL: connection,
		AuthMethod:    authMethod,
		Token:         token,
		Features:      src.GetFeatures(),
	}

	res, err := clients.OnboardRadarSecretManager(ctx, r.client, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Radar secret manager", err.Error())
		return
	}

	src.SetID(types.StringValue(res.GetPayload().ID))
	src.SetProjectID(types.StringValue(projectID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &src)...)
}

func (r *secretManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	src, diags := r.GetSecretManagerFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	res, err := clients.GetRadarSecretManager(ctx, r.client, projectID, src.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar secret manager not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar secret manager", err.Error())
		return
	}

	// Resource is marked as deleted on the server.
	if res.GetPayload().Deleted {
		// Don't update or remove the state, because its has not been fully deleted server side.
		tflog.Warn(ctx, "Radar secret manager marked for deletion.")
		return
	}

	// Read the details for the secret manager features, incase it changed outside of Terraform.
	features := res.GetPayload().Features
	tflog.Warn(ctx, fmt.Sprintf("Read of radar secret manager features: %+v type:%T ", features, features))
	if featuresMap, ok := features.(map[string]interface{}); ok {
		src.SetFeatures(featuresMap)
	} else {
		src.SetFeatures(nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &src)...)
}

func (r *secretManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	src, diags := r.GetSecretManagerFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	// Assert resource still exists.
	res, err := clients.GetRadarSecretManager(ctx, r.client, projectID, src.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar secret manager not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar secret manager", err.Error())
		return
	}

	// Resource is already marked as being deleted on the server. Wait for it to be fully deleted.
	if res.GetPayload().Deleted {
		tflog.Info(ctx, "Radar resource already marked for deletion, waiting for full deletion")
		if err := clients.WaitOnOffboardRadarSecretManager(ctx, r.client, projectID, src.GetID().ValueString()); err != nil {
			resp.Diagnostics.AddError("Unable to delete Radar secret manager", err.Error())
			return
		}

		tflog.Trace(ctx, "Deleted Radar resource")
		return
	}

	// Offboard the Radar secret manager.
	if err := clients.OffboardRadarSecretManager(ctx, r.client, projectID, src.GetID().ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Radar secret manager", err.Error())
		return
	}
}

func (r *secretManagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan, planDiags := r.GetSecretManagerFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(planDiags...)

	state, stateDiags := r.GetSecretManagerFromState(ctx, req.State)
	resp.Diagnostics.Append(stateDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.GetProjectID().IsUnknown() {
		projectID = plan.GetProjectID().ValueString()
	}

	// Check if the features were updated
	planFeatures := plan.GetFeatures()
	stateFeatures := state.GetToken()
	if !reflect.DeepEqual(planFeatures, stateFeatures) {
		body := service.PatchSecretManagerFeaturesBody{
			ID:       plan.GetID().ValueString(),
			Features: planFeatures,
		}

		if err := clients.PatchRadarSecretManagerFeatures(ctx, r.client, projectID, body); err != nil {
			resp.Diagnostics.AddError("Error Updating Radar secret manager features", err.Error())
			return
		}
	}

	// Store the updated plan values
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
