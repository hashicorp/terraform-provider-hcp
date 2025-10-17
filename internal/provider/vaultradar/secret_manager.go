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
// Examples: hcp_resource_radar_secret_manager_vault_dedicated make use of this implementation to define resources with
// specific schemas, validation, and state details related to its type.
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
	GetToken() (types.String, error)
	GetAuthMethod() types.String
	SetFeatures(map[string]interface{})
	GetFeatures(omitEmptyValues bool) map[string]interface{}
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
	sm, diags := r.GetSecretManagerFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !sm.GetProjectID().IsUnknown() {
		projectID = sm.GetProjectID().ValueString()
	}

	connection := sm.GetConnectionURL().ValueString()

	var name string
	if i := strings.Index(sm.GetConnectionURL().ValueString(), ":"); i >= 0 {
		name = connection[:i]
	} else {
		name = connection
	}

	var authMethod string
	if sm.GetAuthMethod().IsNull() || sm.GetAuthMethod().IsUnknown() {
		// This should be caught by schema validation, but just in case.
		resp.Diagnostics.AddError("Error creating Radar secret manager", "missing auth details.")
		return
	}
	authMethod = sm.GetAuthMethod().ValueString()

	tokenValue, err := sm.GetToken()
	if err != nil {
		// This should not happen, but just in case there was an issue constructing the token value.
		resp.Diagnostics.AddError("Error creating Radar secret manager", fmt.Sprintf("unexpected issue with auth details : %s", err.Error()))
		return
	} else if tokenValue.IsNull() || tokenValue.IsUnknown() {
		// This should be caught by schema validation, but just in case.
		resp.Diagnostics.AddError("Error creating Radar secret manager", "auth details must be specified.")
		return
	}
	var token = tokenValue.ValueString()

	// When creating the secret manager, we omit any features that are empty.
	// E.g. features for read-only would be just `{}` where for read-write it would be `{"copy_secrets": {}}`
	const omitEmptyFeatures = true

	body := service.OnboardSecretManagerBody{
		DetectorType:  "agent",
		Type:          r.SecretManagerType,
		Name:          name,
		ConnectionURL: connection,
		AuthMethod:    authMethod,
		Token:         token,
		Features:      sm.GetFeatures(omitEmptyFeatures),
	}

	res, err := clients.OnboardRadarSecretManager(ctx, r.client, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Radar secret manager", err.Error())
		return
	}

	sm.SetID(types.StringValue(res.GetPayload().ID))
	sm.SetProjectID(types.StringValue(projectID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &sm)...)
}

func (r *secretManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	sm, diags := r.GetSecretManagerFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !sm.GetProjectID().IsUnknown() {
		projectID = sm.GetProjectID().ValueString()
	}

	res, err := clients.GetRadarSecretManager(ctx, r.client, projectID, sm.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
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
	tflog.Debug(ctx, fmt.Sprintf("Read of radar secret manager features: %+v type:%T ", features, features))
	if featuresMap, ok := features.(map[string]interface{}); ok {
		sm.SetFeatures(featuresMap)
	} else {
		sm.SetFeatures(nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &sm)...)
}

func (r *secretManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	sm, diags := r.GetSecretManagerFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !sm.GetProjectID().IsUnknown() {
		projectID = sm.GetProjectID().ValueString()
	}

	// Assert the secret manager still exists on the server.
	res, err := clients.GetRadarSecretManager(ctx, r.client, projectID, sm.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Radar secret manager not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar secret manager", err.Error())
		return
	}

	// Already marked as being deleted on the server. Wait for it to be fully deleted.
	if res.GetPayload().Deleted {
		tflog.Info(ctx, "Radar resource already marked for deletion, waiting for full deletion")
		if err := clients.WaitOnOffboardRadarSecretManager(ctx, r.client, projectID, sm.GetID().ValueString()); err != nil {
			resp.Diagnostics.AddError("Unable to delete Radar secret manager", err.Error())
			return
		}

		tflog.Trace(ctx, "Deleted Radar secret manager")
		return
	}

	// Offboard the Radar secret manager.
	if err := clients.OffboardRadarSecretManager(ctx, r.client, projectID, sm.GetID().ValueString()); err != nil {
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

	// Update token ...
	planToken, err := plan.GetToken()
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Radar secret manager auth settings", fmt.Sprintf("unexpected issue with auth details : %s", err.Error()))
		return
	} else if planToken.IsNull() || planToken.IsUnknown() {
		resp.Diagnostics.AddError("Error Updating Radar secret manager auth settings", "auth details must be specified.")
		return
	}

	stateToken, _ := state.GetToken() // error should not happen as state was already read successfully. Besides, if there was an error, we want to update the token.

	if !planToken.Equal(stateToken) {
		body := service.UpdateSecretManagerTokenBody{
			ID:         plan.GetID().ValueString(),
			Token:      planToken.ValueString(),
			AuthMethod: plan.GetAuthMethod().ValueString(),
		}

		if err := clients.UpdateRadarSecretManagerToken(ctx, r.client, projectID, body); err != nil {
			resp.Diagnostics.AddError("Error Updating Radar secret manager auth settings", err.Error())
			return
		}
	}

	// Update features ...
	// When updating the secret manager features, we want to need to include empty values so the service knows what to unset.
	// E.g. features for read-only would be just `{"copy_secrets": null}` where for read-write it would be `{"copy_secrets": {}}`
	const doNotOmitEmptyFeatures = false
	planFeatures := plan.GetFeatures(doNotOmitEmptyFeatures)
	stateFeatures := state.GetFeatures(doNotOmitEmptyFeatures)
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
