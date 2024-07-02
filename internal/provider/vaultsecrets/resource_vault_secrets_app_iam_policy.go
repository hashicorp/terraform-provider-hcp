// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
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
)

// vaultSecretsAppIAMSchema is the schema for the vault secret resource IAM resources
// (policy/binding). It will be merged with the base policy.
func vaultSecretsAppIAMSchema(binding bool) schema.Schema {
	// Determine the description based on if it is for the policy or binding
	d := "Sets the Vault Secrets App IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the Vault Secrets App IAM policy to bind a role to a new member. Existing bindings are preserved."
	}

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				Required:    true,
				Description: "The app's resource name in the format secrets/project/<project ID>/app/<app Name>.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func NewVaultSecretsAppIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("vault_secrets_app", vaultSecretsAppIAMSchema(false), "resource_name", newVaultSecretsAppResourceIAMPolicyUpdater)
}

func NewVaultSecretsAppIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("vault_secrets_app", vaultSecretsAppIAMSchema(true), "resource_name", newVaultSecretsAppResourceIAMPolicyUpdater)
}

type vaultSecretsAppResourceIAMPolicyUpdater struct {
	resourceName string
	client       *clients.Client
	d            iampolicy.TerraformResourceData
}

func newVaultSecretsAppResourceIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	var resourceName types.String
	diags := d.GetAttribute(ctx, path.Root("resource_name"), &resourceName)

	return &vaultSecretsAppResourceIAMPolicyUpdater{
		resourceName: resourceName.ValueString(),
		client:       clients,
		d:            d,
	}, diags
}

func (u *vaultSecretsAppResourceIAMPolicyUpdater) GetMutexKey() string {
	return u.resourceName
}

// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
func (u *vaultSecretsAppResourceIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
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
			return &models.HashicorpCloudResourcemanagerPolicy{}, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to retrieve resource IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// SetResourceIamPolicy Replaces the existing IAM Policy attached to a resource.
func (u *vaultSecretsAppResourceIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceSetIamPolicyParams()

	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		Policy:       policy,
		ResourceName: u.resourceName,
	}

	res, err := u.client.ResourceService.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*resource_service.ResourceServiceSetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast resource IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to update resource IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

var (
	_ iampolicy.NewResourceIamUpdaterFunc = newVaultSecretsAppResourceIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &vaultSecretsAppResourceIAMPolicyUpdater{}
)
