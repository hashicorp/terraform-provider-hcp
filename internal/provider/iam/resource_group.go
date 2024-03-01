package iam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

func NewGroupResource() resource.Resource {
	return &resourceGroup{}
}

type resourceGroup struct {
	client *clients.Client
}

func (r *resourceGroup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *resourceGroup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf(`The group resource manages a HCP Group.

The user or service account that is running Terraform when creating a %s resource must have %s on the parent resource; either the project or organization.`,
			"`hcp_group`", "`roles/admin`"),
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The group's unique identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("The group's resource name in the format `%s` or `%s`",
					"iam/organization/<organization_id>/service-principal/<name>", "<name>"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "The group's display_name.",
				Validators: []validator.String{
					hcpvalidator.ResourceNamePart(),
					// TODO: we don't enforce a list for the display name in the portal, do we care here?
					stringvalidator.LengthBetween(3, 36),
				},
			},
			"description": schema.StringAttribute{
				Description: "The group's description ",
				Optional:    true,
			},
		},
	}
}

func (r *resourceGroup) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type Group struct {
	ResourceID   types.String `tfsdk:"resource_id"`
	ResourceName types.String `tfsdk:"resource_name"`
	DisplayName  types.String `tfsdk:"display_name"`
	Description  types.String `tfsdk:"description"`
}

func (r *resourceGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Group
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.client.Config.OrganizationID
	parent := fmt.Sprintf("organization/%s", orgID)

	createParams := groups_service.NewGroupsServiceCreateGroupParams()
	createParams.ParentResourceName = parent
	createParams.Body = groups_service.GroupsServiceCreateGroupBody{
		Name:        plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	res, err := r.client.Groups.GroupsServiceCreateGroup(createParams, nil)

	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	group := res.GetPayload().Group

	plan.ResourceID = types.StringValue(group.ResourceID)
	plan.ResourceName = types.StringValue(group.ResourceName)
	plan.DisplayName = types.StringValue(group.DisplayName)
	plan.Description = types.StringValue(group.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *resourceGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Group
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	getParams := groups_service.NewGroupsServiceGetGroupParams()
	getParams.ResourceName = state.ResourceName.ValueString()
	res, err := r.client.Groups.GroupsServiceGetGroup(getParams, nil)

	if err != nil {
		var getErr *groups_service.GroupsServiceGetGroupDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error retrieving group", err.Error())
		return
	}

	group := res.GetPayload().Group

	state.ResourceID = types.StringValue(group.ResourceID)
	state.ResourceName = types.StringValue(group.ResourceName)
	state.DisplayName = types.StringValue(group.DisplayName)
	state.Description = types.StringValue(group.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Group
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateParams := groups_service.NewGroupsServiceUpdateGroupParams()
	updateParams.ResourceName = state.ResourceName.ValueString()
	updateParams.Body = groups_service.GroupsServiceUpdateGroupBody{Group: &models.HashicorpCloudIamGroup{ResourceName: state.ResourceName.ValueString(), ResourceID: state.ResourceID.ValueString()}}
	paths := []string{}

	// Check if the display name was updated
	if !plan.DisplayName.Equal(state.DisplayName) {
		updateParams.Body.Group.DisplayName = plan.DisplayName.ValueString()
		paths = append(paths, "displayName")
	}

	// Check if the description was updated
	if !plan.Description.Equal(state.Description) {
		updateParams.Body.Group.Description = plan.Description.ValueString()
		paths = append(paths, "description")
	}

	mask := strings.Join(paths, `,`)
	updateParams.Body.UpdateMask = mask
	_, err := r.client.Groups.GroupsServiceUpdateGroup(updateParams, nil)

	if err != nil {
		resp.Diagnostics.AddError("Error updating group", err.Error())
		return
	}

	// Store the updated values
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Group
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteParams := groups_service.NewGroupsServiceDeleteGroupParams()
	deleteParams.ResourceName = state.ResourceName.ValueString()
	_, err := r.client.Groups.GroupsServiceDeleteGroup(deleteParams, nil)

	if err != nil {
		var getErr *groups_service.GroupsServiceDeleteGroupDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting group", err.Error())
		return
	}
}

func (r *resourceGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_name"), req, resp)
}
