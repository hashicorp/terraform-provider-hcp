// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
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
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/iam/helper"
)

// groupIAMSchema is the schema for the group IAM resources
// (policy/binding). It will be merged with the base policy.
func groupIAMSchema(binding bool) schema.Schema {
	// Determine the description based on if it is for the policy or binding
	d := "Sets the group's IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the group's IAM policy to bind a role to a new member. Existing bindings are preserved."
	}

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: fmt.Sprintf(
					"The group's resource name in format `%s`. The shortened `%s` version can be used for input.",
					"iam/organization/<organization_id>/group/<group_name>",
					"<group_name>",
				),
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func NewGroupIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("group", groupIAMSchema(false), "name", newGroupIAMPolicyUpdater)
}

func NewGroupIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("group", groupIAMSchema(true), "name", newGroupIAMPolicyUpdater)
}

type groupIAMPolicyUpdater struct {
	resourceName string
	client       *clients.Client
	d            iampolicy.TerraformResourceData
}

func newGroupIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	// Determine the resource name
	var resourceName types.String
	diags := d.GetAttribute(ctx, path.Root("name"), &resourceName)

	return &groupIAMPolicyUpdater{
		resourceName: helper.ResourceName(resourceName.ValueString(), clients.Config.OrganizationID),
		client:       clients,
		d:            d,
	}, diags
}

func (u *groupIAMPolicyUpdater) GetMutexKey() string {
	return u.resourceName
}

// Fetch the existing IAM policy attached to a resource.
func (u *groupIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceGetIamPolicyParams()
	params.ResourceName = &u.resourceName

	res, err := u.client.ResourceService.ResourceServiceGetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*resource_service.ResourceServiceGetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast resource IAM policy error", err.Error())
			return nil, diags
		}
		if serviceErr.Code() == http.StatusNotFound {
			// Groups do not have a policy by default
			return &models.HashicorpCloudResourcemanagerPolicy{}, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to retrieve group IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// Replaces the existing IAM Policy attached to a resource.
func (u *groupIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceSetIamPolicyParams()
	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		ResourceName: u.resourceName,
		Policy:       policy,
	}

	res, err := u.client.ResourceService.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*resource_service.ResourceServiceSetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast resource IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to set group IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

var (
	_ iampolicy.NewResourceIamUpdaterFunc = newGroupIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &groupIAMPolicyUpdater{}
)
