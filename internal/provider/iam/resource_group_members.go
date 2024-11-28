// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewGroupMembersResource() resource.Resource {
	return &resourceGroupMembers{}
}

type resourceGroupMembers struct {
	client *clients.Client
}

type GroupMembers struct {
	Group   types.String   `tfsdk:"group"`
	Members []types.String `tfsdk:"members"`
}

func (r *resourceGroupMembers) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (r *resourceGroupMembers) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf(`The group members resource manages the members of an HCP Group.

The user or service account that is running Terraform when creating an %s resource must have %s on the organization.`,
			"`hcp_group_members`", "`roles/admin`"),
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Required: true,
				Description: fmt.Sprintf("The group's resource name in the format `%s`",
					"iam/organization/<organization_id>/group/<name>"),
			},
			"members": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A list of user principal IDs to add to the group.",
			},
		},
	}
}

func (r *resourceGroupMembers) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceGroupMembers) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupMembers
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listParams := groups_service.NewGroupsServiceListGroupMembersParams().WithContext(ctx)
	listParams.SetResourceName(plan.Group.ValueString())
	res, err := r.client.Groups.GroupsServiceListGroupMembers(listParams, nil)
	if err != nil {
		var listResp *groups_service.GroupsServiceListGroupMembersDefault
		if errors.As(err, &listResp) && !listResp.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Failed to list group members", err.Error())
			return
		}
	} else if len(res.GetPayload().Members) > 0 {
		resp.Diagnostics.AddError("Group already has members", "You need to import the resource first.")
		return
	}

	members := make([]string, len(plan.Members))
	for i, member := range plan.Members {
		members[i] = member.ValueString()
	}

	updateParams := groups_service.NewGroupsServiceUpdateGroupMembersParams().WithContext(ctx)
	updateParams.SetResourceName(plan.Group.ValueString())
	updateParams.SetBody(groups_service.GroupsServiceUpdateGroupMembersBody{
		MemberPrincipalIdsToAdd: members,
	})

	_, err = clients.UpdateGroupMembersRetry(r.client, updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update group members", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGroupMembers) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupMembers
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listParams := groups_service.NewGroupsServiceListGroupMembersParams().WithContext(ctx)
	listParams.SetResourceName(state.Group.ValueString())
	res, err := r.client.Groups.GroupsServiceListGroupMembers(listParams, nil)
	if err != nil {
		var listResp *groups_service.GroupsServiceListGroupMembersDefault
		if errors.As(err, &listResp) && listResp.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Failed to list group members", err.Error())
		return
	}

	members := res.GetPayload().Members

	state.Members = make([]types.String, len(members))
	for i, member := range members {
		state.Members[i] = types.StringValue(member.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGroupMembers) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GroupMembers
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planMembers := make(map[string]bool)
	for _, member := range plan.Members {
		planMembers[member.ValueString()] = true
	}

	stateMembers := make(map[string]bool)
	for _, member := range state.Members {
		stateMembers[member.ValueString()] = true
	}

	membersToAdd := make([]string, 0, len(plan.Members))
	for _, member := range plan.Members {
		memberStr := member.ValueString()
		if _, ok := stateMembers[memberStr]; !ok {
			membersToAdd = append(membersToAdd, memberStr)
		}
	}

	membersToRemove := make([]string, 0, len(state.Members))
	for _, member := range state.Members {
		memberStr := member.ValueString()
		if _, ok := planMembers[memberStr]; !ok {
			membersToRemove = append(membersToRemove, memberStr)
		}
	}

	if len(membersToAdd) > 0 || len(membersToRemove) > 0 {
		updateParams := groups_service.NewGroupsServiceUpdateGroupMembersParams().WithContext(ctx)
		updateParams.SetResourceName(plan.Group.ValueString())
		updateParams.SetBody(groups_service.GroupsServiceUpdateGroupMembersBody{
			MemberPrincipalIdsToAdd:    membersToAdd,
			MemberPrincipalIdsToRemove: membersToRemove,
		})

		_, err := clients.UpdateGroupMembersRetry(r.client, updateParams)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update group members", err.Error())
			return
		}
	}

	// Store the updated values
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGroupMembers) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupMembers
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	members := make([]string, len(state.Members))
	for i, member := range state.Members {
		members[i] = member.ValueString()
	}

	updateParams := groups_service.NewGroupsServiceUpdateGroupMembersParams().WithContext(ctx)
	updateParams.SetResourceName(state.Group.ValueString())
	updateParams.SetBody(groups_service.GroupsServiceUpdateGroupMembersBody{
		MemberPrincipalIdsToRemove: members,
	})

	_, err := clients.UpdateGroupMembersRetry(r.client, updateParams)
	if err != nil {
		var errResp *groups_service.GroupsServiceUpdateGroupMembersDefault
		if errors.As(err, &errResp) && !errResp.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Failed to update group members", err.Error())
			return
		}
	}
}

func (r *resourceGroupMembers) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group"), req, resp)
}
