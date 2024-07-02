// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/iampolicy"
	"github.com/hashicorp/terraform-provider-hcp/internal/customdiags"
)

// projectIAMSchema is the schema for the project IAM resources
// (policy/binding). It will be merged with the base policy.
func projectIAMSchema(binding bool) schema.Schema {
	// Determine the description based on if it is for the policy or binding
	d := "Sets the project's IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the project's IAM policy to bind a role to a new member. Existing bindings are preserved."
	}

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project to apply the IAM Policy to. If unspecified, the project configured on the provider is used.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func NewProjectIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("project", projectIAMSchema(false), "project_id", newProjectIAMPolicyUpdater)
}

func NewProjectIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("project", projectIAMSchema(true), "project_id", newProjectIAMPolicyUpdater)
}

type projectIAMPolicyUpdater struct {
	projectID string
	client    *clients.Client
	d         iampolicy.TerraformResourceData
}

func newProjectIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	// Determine the project ID
	var projectID types.String
	diags := d.GetAttribute(ctx, path.Root("project_id"), &projectID)
	if projectID.ValueString() == "" {
		projectID = types.StringValue(clients.Config.ProjectID)
	}

	return &projectIAMPolicyUpdater{
		projectID: projectID.ValueString(),
		client:    clients,
		d:         d,
	}, diags
}

func (u *projectIAMPolicyUpdater) GetMutexKey() string {
	return u.projectID
}

// Fetch the existing IAM policy attached to a resource.
func (u *projectIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := project_service.NewProjectServiceGetIamPolicyParams()
	params.ID = u.projectID
	res, err := u.client.Project.ProjectServiceGetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*project_service.ProjectServiceGetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast project IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to retrieve project IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// Replaces the existing IAM Policy attached to a resource.
func (u *projectIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := project_service.NewProjectServiceSetIamPolicyParams()
	params.ID = u.projectID
	params.Body = project_service.ProjectServiceSetIamPolicyBody{
		Policy: policy,
	}

	res, err := u.client.Project.ProjectServiceSetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*project_service.ProjectServiceSetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast project IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to update project IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

var (
	_ iampolicy.NewResourceIamUpdaterFunc = newProjectIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &projectIAMPolicyUpdater{}
)
