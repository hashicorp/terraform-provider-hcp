// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"

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
)

// hvsIAMSchema is the schema for the HVS resource IAM resources
// (policy/binding). It will be merged with the base policy.
func hvsIAMSchema(binding bool) schema.Schema {
	// Determine the description based on if it is for the policy or binding
	d := "Sets the vault secrets app IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the vault secrets app IAM policy to bind a role to a new member. Existing bindings are preserved."
	}

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the HCP resource to apply the IAM Policy to. Either this of the resource name must be provided",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_name": schema.StringAttribute{
				Optional:    true,
				Description: "The Name of the HCP resource to apply the IAM Policy to. Either this of the resource ID must be provided",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func NewHVSIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("vault_secrets_app", hvsIAMSchema(false), "resource_id", newHVSResourceIAMPolicyUpdater)
}

func NewHVSIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("vault_secrets_app", hvsIAMSchema(true), "resource_id", newHVSResourceIAMPolicyUpdater)
}

type hvsResourceIAMPolicyUpdater struct {
	resourceID   string
	resourceName string
	client       *clients.Client
	d            iampolicy.TerraformResourceData
}

func newHVSResourceIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	// Determine the resource ID and resource name
	var resourceID, resourceName types.String
	diags := d.GetAttribute(ctx, path.Root("resource_id"), &resourceID)
	diags = d.GetAttribute(ctx, path.Root("resource_name"), &resourceName)
	if resourceID.ValueString() == "" && resourceName.ValueString() == "" {
		diags.AddError("missing resource ID and resource Name", "resource ID and resource Name are both missing. One of these must be provided")
		return nil, diags
	}

	if resourceID.ValueString() != "" && resourceName.ValueString() != "" {
		diags.AddError("both resource ID and resource Name provided", "resource ID and resource Name are both present. Only ne of these must be provided")
		return nil, diags
	}

	return &hvsResourceIAMPolicyUpdater{
		resourceID:   resourceID.ValueString(),
		resourceName: resourceName.ValueString(),
		client:       clients,
		d:            d,
	}, diags
}

func (u *hvsResourceIAMPolicyUpdater) GetMutexKey() string {
	return u.resourceID
}

// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
func (u *hvsResourceIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceGetIamPolicyParams()
	if u.resourceID != "" {
		params.ResourceID = &u.resourceID
	} else {
		params.ResourceName = &u.resourceName
	}
	res, err := u.client.ResourceService.ResourceServiceGetIamPolicy(params, nil)
	if err != nil {
		diags.AddError("failed to retrieve resource IAM policy", err.Error())
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// SetResourceIamPolicy Replaces the existing IAM Policy attached to a resource.
func (u *hvsResourceIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceSetIamPolicyParams()

	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		Policy: policy,
	}
	if u.resourceID != "" {
		params.Body.ResourceID = u.resourceID
	} else {
		params.Body.ResourceName = u.resourceName
	}

	res, err := u.client.ResourceService.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		diags.AddError("failed to retrieve resource IAM policy", err.Error())
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

var (
	_ iampolicy.NewResourceIamUpdaterFunc = newHVSResourceIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &hvsResourceIAMPolicyUpdater{}
)
