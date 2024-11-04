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

	service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_connection_service"
)

var (
	_ resource.Resource              = &integrationConnectionResource{}
	_ resource.ResourceWithConfigure = &integrationConnectionResource{}
)

// integrationConnectionResource is an implementation for configuring specific types of integration connections.
// Examples: hcp_vault_radar_integration_jira_connection and hcp_vault_radar_integration_slack_connection make use of
// this implementation to define resources with specific schemas, validation, and state details related to their types.
type integrationConnectionResource struct {
	client                 *clients.Client
	TypeName               string
	IntegrationType        string
	ConnectionSchema       schema.Schema
	GetConnectionFromPlan  func(ctx context.Context, plan tfsdk.Plan) (integrationConnection, diag.Diagnostics)
	GetConnectionFromState func(ctx context.Context, state tfsdk.State) (integrationConnection, diag.Diagnostics)
}

// integrationConnection is the minimal plan/state that a connection must have.
// Specifics to the type of connection should use the GetAuthKey and Get/Set Details for specific plan and state.
type integrationConnection interface {
	GetID() types.String
	SetID(types.String)
	GetProjectID() types.String
	SetProjectID(types.String)
	GetName() types.String
	SetName(types.String)
	GetAuthKey() (string, diag.Diagnostics)
	GetDetails() (string, diag.Diagnostics)
	SetDetails(string) diag.Diagnostics
}

func (r *integrationConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.TypeName
}

func (r *integrationConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.ConnectionSchema
}

func (r *integrationConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationConnectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *integrationConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn, diags := r.GetConnectionFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !conn.GetProjectID().IsUnknown() {
		projectID = conn.GetProjectID().ValueString()
	}

	errSummary := "Error creating Radar Integration Connection"

	// Check for an existing connection with the same name.
	existing, err := clients.GetIntegrationConnectionByName(ctx, r.client, projectID, conn.GetName().ValueString())
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		resp.Diagnostics.AddError(errSummary, err.Error())
	}
	if existing != nil {
		resp.Diagnostics.AddError(errSummary, fmt.Sprintf("Connection with name: %q already exists.", conn.GetName().ValueString()))
		return
	}

	authKey, diags := conn.GetAuthKey()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, diags := conn.GetDetails()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := clients.CreateIntegrationConnection(ctx, r.client, projectID, service.CreateIntegrationConnectionBody{
		IntegrationType: r.IntegrationType,
		IsSink:          true,
		IsSource:        false,
		Name:            conn.GetName().ValueString(),
		AuthKey:         authKey,
		Details:         details,
	})
	if err != nil {
		resp.Diagnostics.AddError(errSummary, err.Error())
		return
	}

	conn.SetID(types.StringValue(res.GetPayload().ID))
	conn.SetProjectID(types.StringValue(projectID))
	resp.Diagnostics.Append(resp.State.Set(ctx, conn)...)
}

func (r *integrationConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn, diags := r.GetConnectionFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !conn.GetProjectID().IsUnknown() {
		projectID = conn.GetProjectID().ValueString()
	}

	res, err := clients.GetIntegrationConnectionByID(ctx, r.client, projectID, conn.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar integration connection not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar integration connection", err.Error())
		return
	}

	conn.SetName(types.StringValue(res.GetPayload().Name))
	diags = conn.SetDetails(res.GetPayload().Details)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &conn)...)
}

func (r *integrationConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn, diags := r.GetConnectionFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !conn.GetProjectID().IsUnknown() {
		projectID = conn.GetProjectID().ValueString()
	}

	// Assert resource still exists.
	if _, err := clients.GetIntegrationConnectionByID(ctx, r.client, projectID, conn.GetID().ValueString()); err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar integration connection not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar integration connection", err.Error())
		return
	}

	if err := clients.DeleteIntegrationConnection(ctx, r.client, projectID, conn.GetID().ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Radar integration connection", err.Error())
		return
	}
}

func (r *integrationConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported.
	// Plans to support updating the token will be in a future iteration.
	resp.Diagnostics.AddError("Unexpected provider error", "This is an internal error, please report this issue to the provider developers")
}
