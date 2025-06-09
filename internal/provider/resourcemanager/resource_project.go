// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	billing "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"
	billingModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

const (
	// defaultBillingAccountID is the ID of the default/only billing account.
	defaultBillingAccountID = "default-account"
)

func NewProjectResource() resource.Resource {
	return &resourceProject{}
}

type resourceProject struct {
	client *clients.Client
}

func (r *resourceProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *resourceProject) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf(`The project resource manages a HCP Project.

The user or service account that is running Terraform when creating a %s resource must have %s on the specified organization.`,
			"`hcp_project`", "`roles/admin`"),
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project's unique identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				Computed:    true,
				Description: "The project's resource name in format \"project/<resource_id>\"",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The project's name.",
				Validators: []validator.String{
					hcpvalidator.DisplayName(),
					stringvalidator.LengthBetween(3, 36),
				},
			},
			"description": schema.StringAttribute{
				Description: "The project's description",
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
				Default: stringdefault.StaticString(""),
			},
		},
	}
}

func (r *resourceProject) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type Project struct {
	ResourceID   types.String `tfsdk:"resource_id"`
	ResourceName types.String `tfsdk:"resource_name"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
}

func (r *resourceProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION
	createParams := project_service.NewProjectServiceCreateParams()
	createParams.Body = &models.HashicorpCloudResourcemanagerProjectCreateRequest{
		Description: plan.Description.ValueString(),
		Name:        plan.Name.ValueString(),
		Parent: &models.HashicorpCloudResourcemanagerResourceID{
			ID:   r.client.Config.OrganizationID,
			Type: &parentType,
		},
	}

	res, err := clients.CreateProjectWithRetry(r.client, createParams)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	// Wait for the project to be created, if an operation ID is returned.
	if res.Payload.OperationID != "" {
		err = waitForProjectOperation(ctx, r.client, "create project", res.Payload.Project.ID, res.Payload.OperationID)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for project creation", err.Error())
			return
		}
	}

	p := res.GetPayload().Project
	plan.ResourceID = types.StringValue(p.ID)
	plan.ResourceName = types.StringValue(fmt.Sprintf("project/%s", p.ID))
	plan.Description = types.StringValue(p.Description)
	plan.Name = types.StringValue(p.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// Set up billing for the project
	if err := r.addToBillingAccount(ctx, p.ID); err != nil {
		resp.Diagnostics.AddError("Error setting up billing for created project", err.Error())
		return
	}
}

func (r *resourceProject) addToBillingAccount(ctx context.Context, projectID string) error {

	req := billing.NewBillingAccountServiceGetParams()
	req.OrganizationID = r.client.Config.OrganizationID
	req.ID = defaultBillingAccountID
	resp, err := r.client.Billing.BillingAccountServiceGet(req, nil)
	if err != nil {
		return fmt.Errorf("listing billing accounts failed: %v", err.Error())
	}

	// Update the BA to include the new project ID
	ba := resp.Payload.BillingAccount
	updateReq := billing.NewBillingAccountServiceUpdateParams()
	updateReq.OrganizationID = ba.OrganizationID
	updateReq.ID = ba.ID
	updateReq.Body = &billingModels.BillingAccountServiceUpdateBody{
		ProjectIds: ba.ProjectIds,
		Name:       ba.Name,
		Country:    ba.Country,
	}

	updateReq.Body.ProjectIds = append(updateReq.Body.ProjectIds, projectID)

	_, err = clients.RetryBillingServiceUpdate(r.client, updateReq)
	if err != nil {
		return fmt.Errorf("updating billing account failed: %v", err.Error())
	}

	return nil
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := project_service.NewProjectServiceGetParams()
	getParams.ID = state.ResourceID.ValueString()
	res, err := r.client.Project.ProjectServiceGet(getParams, nil)
	if err != nil {
		var getErr *project_service.ProjectServiceGetDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error retrieving project", err.Error())
		return
	}

	p := res.GetPayload().Project
	state.Description = types.StringValue(p.Description)
	state.Name = types.StringValue(p.Name)
	state.ResourceName = types.StringValue(fmt.Sprintf("project/%s", p.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the name was updated
	if !plan.Name.Equal(state.Name) {
		setNameReq := project_service.NewProjectServiceSetNameParams()
		setNameReq.ID = plan.ResourceID.ValueString()
		setNameReq.Body = project_service.ProjectServiceSetNameBody{
			Name: plan.Name.ValueString(),
		}

		res, err := clients.SetProjectNameWithRetry(r.client, setNameReq)
		if err != nil {
			resp.Diagnostics.AddError("Error updating project name", err.Error())
			return
		}

		// Wait for the project name to be updated, if an operation ID is returned.
		if res.Payload.OperationID != "" {
			// Wait for the project name to be updated.
			err = waitForProjectOperation(ctx, r.client, "update project name", plan.ResourceID.ValueString(), res.Payload.OperationID)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for project name update", err.Error())
				return
			}
		}
	}

	// Check if the description was updated
	if !plan.Description.Equal(state.Description) {
		setDescReq := project_service.NewProjectServiceSetDescriptionParams()
		setDescReq.ID = plan.ResourceID.ValueString()
		setDescReq.Body = project_service.ProjectServiceSetDescriptionBody{
			Description: plan.Description.ValueString(),
		}

		res, err := clients.SetProjectDescriptionWithRetry(r.client, setDescReq)
		if err != nil {
			resp.Diagnostics.AddError("Error updating project description", err.Error())
			return
		}

		// Wait for the project description to be updated, if an operation ID is returned.
		if res.Payload.OperationID != "" {
			err = waitForProjectOperation(ctx, r.client, "update project description", plan.ResourceID.ValueString(), res.Payload.OperationID)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for project description update", err.Error())
				return
			}
		}

	}

	// Store the updated values
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := project_service.NewProjectServiceDeleteParams()
	getParams.ID = state.ResourceID.ValueString()

	res, err := r.client.Project.ProjectServiceDelete(getParams, nil)
	if err != nil {
		var deleteErr *project_service.ProjectServiceDeleteDefault
		if errors.As(err, &deleteErr) && deleteErr.IsCode(http.StatusNotFound) {
			return
		}

		resp.Diagnostics.AddError("Error deleting project", err.Error())
		return
	}

	// Wait for the project to be deleted, if an operation ID is returned.
	if res.Payload.Operation.ID != "" {
		// For delete operations, the operation is scoped at the organization level
		projectID := ""
		err = waitForProjectOperation(ctx, r.client, "delete project", projectID, res.Payload.Operation.ID)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for project deletion", err.Error())
			return
		}
	}
}

func (r *resourceProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func waitForProjectOperation(ctx context.Context, client *clients.Client, operationName, projectID string, operationID string) error {
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	return clients.WaitForOperation(ctx, client, operationName, loc, operationID)
}
