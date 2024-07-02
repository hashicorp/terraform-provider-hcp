// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/iampolicy"
	"github.com/hashicorp/terraform-provider-hcp/internal/customdiags"
)

// orgIAMSchema is the schema for the organization IAM resources
// (policy/binding). It will be merged with the base policy.
func orgIAMSchema(binding bool) schema.Schema {
	// Determine the description based on if it is for the policy or binding
	d := "Sets the organization's IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the organization's IAM policy to bind a role to a new member. Existing bindings are preserved."
	}

	return schema.Schema{
		MarkdownDescription: d,
	}
}

func NewOrganizationIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("organization", orgIAMSchema(false), "", newOrgIAMPolicyUpdater)
}

func NewOrganizationIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("organization", orgIAMSchema(true), "", newOrgIAMPolicyUpdater)
}

type orgIAMPolicyUpdater struct {
	client *clients.Client
	d      iampolicy.TerraformResourceData
}

func newOrgIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &orgIAMPolicyUpdater{
		client: clients,
		d:      d,
	}, diags
}

func (u *orgIAMPolicyUpdater) GetMutexKey() string {
	return u.client.Config.OrganizationID
}

// Fetch the existing IAM policy attached to a resource.
func (u *orgIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := organization_service.NewOrganizationServiceGetIamPolicyParams()
	params.ID = u.client.Config.OrganizationID
	res, err := u.client.Organization.OrganizationServiceGetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*organization_service.OrganizationServiceGetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast organization IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to retrieve organization IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// Replaces the existing IAM Policy attached to a resource.
func (u *orgIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := organization_service.NewOrganizationServiceSetIamPolicyParams()
	params.ID = u.client.Config.OrganizationID
	params.Body = organization_service.OrganizationServiceSetIamPolicyBody{
		Policy: policy,
	}

	res, err := u.client.Organization.OrganizationServiceSetIamPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*organization_service.OrganizationServiceSetIamPolicyDefault)
		if !ok {
			diags.AddError("failed to cast organization IAM policy error", err.Error())
			return nil, diags
		}
		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to update organization IAM policy", err.Error(), serviceErr.Code()))
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

var (
	_ iampolicy.NewResourceIamUpdaterFunc = newOrgIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &orgIAMPolicyUpdater{}
)
