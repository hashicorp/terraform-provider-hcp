// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ActionConfigResource{}
var _ resource.ResourceWithImportState = &ActionConfigResource{}

type ActionConfigResource struct {
	client *clients.Client
}

// ActionConfigModel describes the resource data model.
type ActionConfigResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
	Description types.String `tfsdk:"description"`
	NamespaceID types.String `tfsdk:"namespace_id"`
	ActionURL   types.String `tfsdk:"action_url"`
	//CreatedAt   types.String `tfsdk:"created_at"`

	Request actionConfigRequest `tfsdk:"request"`
}

type actionConfigRequest struct {
	custom customRequest `tfsdk:"custom"`
}

type customRequest struct {
	Method  types.String `tfsdk:"method"`
	Headers types.Map    `tfsdk:"headers"`
	URL     types.String `tfsdk:"url"`
	Body    types.String `tfsdk:"body"`
}

func ConvertMethodToStringType(method waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethod) (types.String, error) {
	var methodString types.String
	switch method {
	case waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodGET:
		methodString = types.StringValue("GET")
	case waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPOST:
		methodString = types.StringValue("POST")
	case waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPUT:
		methodString = types.StringValue("PUT")
	case waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPATCH:
		methodString = types.StringValue("PATCH")
	case waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodDELETE:
		methodString = types.StringValue("DELETE")
	default:
		return methodString, fmt.Errorf("unknown method")
	}
	return methodString, nil
}

func ConvertMethodToEnumType(method string) (waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethod, error) {
	var methodEnum waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethod
	switch method {
	case "GET":
		methodEnum = waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodGET
	case "POST":
		methodEnum = waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPOST
	case "PUT":
		methodEnum = waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPUT
	case "PATCH":
		methodEnum = waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodPATCH
	case "DELETE":
		methodEnum = waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomMethodDELETE
	default:
		return methodEnum, fmt.Errorf("unknown method")
	}
	return methodEnum, nil
}

func (r *ActionConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_action_config"
}

func (r *ActionConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Action Config resource managed the lifecycle of an Action Config.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Action Config.",
				Computed:    true,
			},
			// TODO: Should Name be optional?
			"name": schema.StringAttribute{
				Description: "The name of the Action Config.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Action Config is located.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Action Config is located.",
				Computed:    true,
			},
			// An Action Config description must be fewer than 125 characters if set.
			"description": schema.StringAttribute{
				Description: "A description of the Action Config.",
				// TODO: Should this be optional?
				Optional: true,
			},
			"namespace_id": schema.StringAttribute{
				Description: "Internal Namespace ID",
				Computed:    true,
			},
			// Action URL is required if the action is custom
			"action_url": schema.StringAttribute{
				Description: "The URL to trigger an action on. Only used in Custom mode",
				Optional:    true,
			},
			/*"created_at": schema.StringAttribute{
				Description: "The timestamp when the Action Config was created in the database.",
				Computed:    true,
			},*/
			"request": schema.ListNestedAttribute{
				Description: "The kind of HTTP request this config should trigger.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"custom": schema.ListNestedAttribute{
							Description: "Custom mode allows users to define the HTTP method, the request body, etc.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"method": schema.StringAttribute{
										Description: "The HTTP method to use for the request.",
										Required:    true,
									},
									"headers": schema.MapAttribute{
										Description: "Key value headers to send with the request.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"url": schema.StringAttribute{
										Description: "The full URL this request should make when invoked.",
										Optional:    true,
									},
									"body": schema.StringAttribute{
										Description: "The body to be submitted with the request.",
										Optional:    true,
									},
								},
							},
						},
						/*"github": schema.ListNestedAttribute{
							Description: "GitHub mode is configured to do various operations on GitHub Repositories.",
							Optional:    true,
						},
						"agent": schema.ListNestedAttribute{
							Optional: true,
						},*/
					},
				},
			},
		},
	}
}

func (r *ActionConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// TODO: Add support for request and created at
func (r *ActionConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ActionConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting namespace by location",
			err.Error(),
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointActionConfig{
			Request: &waypoint_models.HashicorpCloudWaypointActionConfigRequest{},
		},
	}

	if !plan.Name.IsUnknown() {
		modelBody.ActionConfig.Name = plan.Name.ValueString()
	}

	if !plan.Description.IsUnknown() {
		modelBody.ActionConfig.Description = plan.Description.ValueString()
	}

	if !plan.ActionURL.IsUnknown() {
		modelBody.ActionConfig.ActionURL = plan.ActionURL.ValueString()
	}

	// This is a proxy for the request type, as custom.Method is required for custom requests
	if !plan.Request.custom.Method.IsUnknown() {
		modelBody.ActionConfig.Request.Custom = &waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustom{}

		method, err := ConvertMethodToEnumType(plan.Request.custom.Method.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unexpected HTTP Method",
				"Expected GET, POST, PUT, DELETE, or PATCH",
			)
			return
		}
		modelBody.ActionConfig.Request.Custom.Method = &method

		if !plan.Request.custom.Headers.IsUnknown() {
			for key, value := range plan.Request.custom.Headers.Elements() {
				modelBody.ActionConfig.Request.Custom.Headers = append(modelBody.ActionConfig.Request.Custom.Headers, &waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomHeader{
					Key:   key,
					Value: value.String(),
				})
			}
		}
		if !plan.Request.custom.URL.IsUnknown() {
			modelBody.ActionConfig.Request.Custom.URL = plan.Request.custom.URL.ValueString()
		}
		if !plan.Request.custom.Body.IsUnknown() {
			modelBody.ActionConfig.Request.Custom.Body = plan.Request.custom.Body.ValueString()

		}
	}

	params := &waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}

	aCfg, err := r.client.Waypoint.WaypointServiceCreateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Action Config", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointActionConfig
	if aCfg.Payload != nil {
		aCfgModel = aCfg.Payload.ActionConfig
	}
	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error creating Action Config", "Empty Action Config returned")
		return
	}

	plan.ID = types.StringValue(aCfgModel.ID)
	plan.Name = types.StringValue(aCfgModel.Name)
	plan.Description = types.StringValue(aCfgModel.Description)
	plan.ActionURL = types.StringValue(aCfgModel.ActionURL)

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.NamespaceID = types.StringValue(ns.ID)

	plan.Request = actionConfigRequest{}
	headerMap := make(map[types.String][]types.String)

	var diags diag.Diagnostics

	// In the future, expand this to accommodate other types of requests
	if aCfgModel.Request.Custom != nil {
		plan.Request.custom = customRequest{}
		if aCfgModel.Request.Custom.Method != nil {
			methodString, err := ConvertMethodToStringType(*aCfgModel.Request.Custom.Method)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unexpected HTTP Method",
					"Expected GET, POST, PUT, DELETE, or PATCH. Please report this issue to the provider developers.",
				)
			} else {
				plan.Request.custom.Method = methodString
			}
		}
		if aCfgModel.Request.Custom.Headers != nil {
			for _, header := range aCfgModel.Request.Custom.Headers {
				headerMap[types.StringValue(header.Key)] = append(headerMap[types.StringValue(header.Key)], types.StringValue(header.Value))
			}
			plan.Request.custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if aCfgModel.Request.Custom.URL != "" {
			plan.Request.custom.URL = types.StringValue(aCfgModel.Request.Custom.URL)
		}
		if aCfgModel.Request.Custom.Body != "" {
			plan.Request.custom.Body = types.StringValue(aCfgModel.Request.Custom.Body)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created Action Config resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client

	actionCfg, err := clients.GetActionConfig(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action Config not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Action Config", err.Error())
		return
	}

	data.ID = types.StringValue(actionCfg.ID)
	data.Name = types.StringValue(actionCfg.Name)
	data.Description = types.StringValue(actionCfg.Description)
	data.ActionURL = types.StringValue(actionCfg.ActionURL)

	data.ProjectID = types.StringValue(projectID)
	data.OrgID = types.StringValue(orgID)

	data.Request = actionConfigRequest{}
	headerMap := make(map[types.String][]types.String)

	var diags diag.Diagnostics

	// In the future, expand this to accommodate other types of requests
	if actionCfg.Request.Custom != nil {
		data.Request.custom = customRequest{}
		if actionCfg.Request.Custom.Method != nil {
			methodString, err := ConvertMethodToStringType(*actionCfg.Request.Custom.Method)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unexpected HTTP Method",
					"Expected GET, POST, PUT, DELETE, or PATCH. Please report this issue to the provider developers.",
				)
			} else {
				data.Request.custom.Method = methodString
			}
		}
		if actionCfg.Request.Custom.Headers != nil {
			for _, header := range actionCfg.Request.Custom.Headers {
				headerMap[types.StringValue(header.Key)] = append(headerMap[types.StringValue(header.Key)], types.StringValue(header.Value))
			}
			data.Request.custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if actionCfg.Request.Custom.URL != "" {
			data.Request.custom.URL = types.StringValue(actionCfg.Request.Custom.URL)
		}
		if actionCfg.Request.Custom.Body != "" {
			data.Request.custom.Body = types.StringValue(actionCfg.Request.Custom.Body)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ActionConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get the current state as well, so we know the current name of the
	// action config for reference during the update
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting namespace by location",
			err.Error(),
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointActionConfig{
			Request: &waypoint_models.HashicorpCloudWaypointActionConfigRequest{},
		},
	}

	// These are the updated values
	if !plan.Name.IsUnknown() {
		modelBody.ActionConfig.Name = plan.Name.ValueString()
	}
	if !plan.Description.IsUnknown() {
		modelBody.ActionConfig.Description = plan.Description.ValueString()
	}
	if !plan.ActionURL.IsUnknown() {
		modelBody.ActionConfig.ActionURL = plan.ActionURL.ValueString()
	}

	// This is a proxy for the request type, as custom.Method is required for custom requests
	if !plan.Request.custom.Method.IsUnknown() {
		modelBody.ActionConfig.Request.Custom = &waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustom{}

		method, err := ConvertMethodToEnumType(plan.Request.custom.Method.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unexpected HTTP Method",
				"Expected GET, POST, PUT, DELETE, or PATCH",
			)
			return
		}
		modelBody.ActionConfig.Request.Custom.Method = &method

		if !plan.Request.custom.Headers.IsUnknown() {
			for key, value := range plan.Request.custom.Headers.Elements() {
				modelBody.ActionConfig.Request.Custom.Headers = append(modelBody.ActionConfig.Request.Custom.Headers, &waypoint_models.HashicorpCloudWaypointActionConfigFlavorCustomHeader{
					Key:   key,
					Value: value.String(),
				})
			}
		}
		if !plan.Request.custom.URL.IsUnknown() {
			modelBody.ActionConfig.Request.Custom.URL = plan.Request.custom.URL.ValueString()
		}
		if !plan.Request.custom.Body.IsUnknown() {
			modelBody.ActionConfig.Request.Custom.Body = plan.Request.custom.Body.ValueString()

		}
	}

	params := &waypoint_service.WaypointServiceUpdateActionConfigParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}

	actionCfg, err := r.client.Waypoint.WaypointServiceUpdateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Action Config", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointActionConfig
	if actionCfg.Payload != nil {
		aCfgModel = actionCfg.Payload.ActionConfig
	}
	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error updating Action Config", "Empty Action Config returned")
		return
	}

	plan.ID = types.StringValue(aCfgModel.ID)
	plan.Name = types.StringValue(aCfgModel.Name)
	plan.Description = types.StringValue(aCfgModel.Description)
	plan.ActionURL = types.StringValue(aCfgModel.ActionURL)

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.NamespaceID = types.StringValue(ns.ID)

	plan.Request = actionConfigRequest{}
	headerMap := make(map[types.String][]types.String)

	var diags diag.Diagnostics

	// In the future, expand this to accommodate other types of requests
	if aCfgModel.Request.Custom != nil {
		plan.Request.custom = customRequest{}
		if aCfgModel.Request.Custom.Method != nil {
			methodString, err := ConvertMethodToStringType(*aCfgModel.Request.Custom.Method)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unexpected HTTP Method",
					"Expected GET, POST, PUT, DELETE, or PATCH. Please report this issue to the provider developers.",
				)
			} else {
				plan.Request.custom.Method = methodString
			}
		}
		if aCfgModel.Request.Custom.Headers != nil {
			for _, header := range aCfgModel.Request.Custom.Headers {
				headerMap[types.StringValue(header.Key)] = append(headerMap[types.StringValue(header.Key)], types.StringValue(header.Value))
			}
			plan.Request.custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if aCfgModel.Request.Custom.URL != "" {
			plan.Request.custom.URL = types.StringValue(aCfgModel.Request.Custom.URL)
		}
		if aCfgModel.Request.Custom.Body != "" {
			plan.Request.custom.Body = types.StringValue(aCfgModel.Request.Custom.Body)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Updated Action Config resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Action Config",
			err.Error(),
		)
		return
	}

	params := &waypoint_service.WaypointServiceDeleteActionConfigParams{
		NamespaceID: ns.ID,
		ActionID:    data.ID.ValueStringPointer(),
		ActionName:  data.Name.ValueStringPointer(),
	}

	_, err = r.client.Waypoint.WaypointServiceDeleteActionConfig(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action Config not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Action Config",
			err.Error(),
		)
		return
	}
}
func (r *ActionConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
