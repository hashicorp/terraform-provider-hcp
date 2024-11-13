// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"

	service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_subscription_service"
)

var (
	_ resource.Resource              = &integrationSubscriptionResource{}
	_ resource.ResourceWithConfigure = &integrationSubscriptionResource{}
)

// integrationSubscriptionResource is an implementation for configuring specific types of integration subscriptions.
// Examples: hcp_vault_radar_integration_jira_subscription and hcp_vault_radar_integration_slack_subscription make use of
// this implementation to define resources with specific schemas, validation, and state details related to their types.
type integrationSubscriptionResource struct {
	client                   *clients.Client
	TypeName                 string
	SubscriptionSchema       schema.Schema
	GetSubscriptionFromPlan  func(ctx context.Context, plan tfsdk.Plan) (integrationSubscription, diag.Diagnostics)
	GetSubscriptionFromState func(ctx context.Context, state tfsdk.State) (integrationSubscription, diag.Diagnostics)
}

// integrationSubscription is the minimal plan/state that a subscription must have.
// Specifics to the type of subscription should use the Get/Set Details for specific plan and state.
type integrationSubscription interface {
	GetID() types.String
	SetID(types.String)
	GetProjectID() types.String
	SetProjectID(types.String)
	GetName() types.String
	SetName(types.String)
	GetConnectionID() types.String
	SetConnectionID(types.String)
	GetDetails() (string, diag.Diagnostics)
	SetDetails(string) diag.Diagnostics
}

func (r *integrationSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.TypeName
}

func (r *integrationSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.SubscriptionSchema
}

func (r *integrationSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationSubscriptionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *integrationSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	subscription, diags := r.GetSubscriptionFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !subscription.GetProjectID().IsUnknown() {
		projectID = subscription.GetProjectID().ValueString()
	}

	errSummary := "Error creating Radar Integration Subscription"

	// Check for an existing subscription with the same name.
	existing, err := clients.GetIntegrationSubscriptionByName(ctx, r.client, projectID, subscription.GetName().ValueString())
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		resp.Diagnostics.AddError(errSummary, err.Error())
	}
	if existing != nil {
		resp.Diagnostics.AddError(errSummary, fmt.Sprintf("Subscription with name: %q already exists.", subscription.GetName().ValueString()))
		return
	}

	details, diags := subscription.GetDetails()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := clients.CreateIntegrationSubscription(ctx, r.client, projectID, service.CreateIntegrationSubscriptionBody{
		Name:         subscription.GetName().ValueString(),
		ConnectionID: subscription.GetConnectionID().ValueString(),
		Details:      details,
	})
	if err != nil {
		resp.Diagnostics.AddError(errSummary, err.Error())
		return
	}

	subscription.SetID(types.StringValue(res.GetPayload().ID))
	subscription.SetProjectID(types.StringValue(projectID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &subscription)...)
}

func (r *integrationSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	subscription, diags := r.GetSubscriptionFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !subscription.GetProjectID().IsUnknown() {
		projectID = subscription.GetProjectID().ValueString()
	}

	res, err := clients.GetIntegrationSubscriptionByID(ctx, r.client, projectID, subscription.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar integration subscription not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar integration subscription", err.Error())
		return
	}

	payload := res.GetPayload()
	subscription.SetName(types.StringValue(payload.Name))
	subscription.SetConnectionID(types.StringValue(payload.ConnectionID))

	diags = subscription.SetDetails(res.GetPayload().Details)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &subscription)...)
}

func (r *integrationSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	subscription, diags := r.GetSubscriptionFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !subscription.GetProjectID().IsUnknown() {
		projectID = subscription.GetProjectID().ValueString()
	}

	// Assert resource still exists.
	if _, err := clients.GetIntegrationSubscriptionByID(ctx, r.client, projectID, subscription.GetID().ValueString()); err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar integration subscription not found, removing from subscription.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar integration subscription", err.Error())
		return
	}

	if err := clients.DeleteIntegrationSubscription(ctx, r.client, projectID, subscription.GetID().ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Radar integration subscription", err.Error())
		return
	}
}

func (r *integrationSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported.
	// Plans to support updating subscription details will be in a future iteration.
	resp.Diagnostics.AddError("Unexpected provider error", "This is an internal error, please report this issue to the provider developers")
}
