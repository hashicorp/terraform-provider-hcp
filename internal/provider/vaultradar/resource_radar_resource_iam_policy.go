// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"

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

// radarResourceIAMSchema is the schema for the vault radar resource IAM resources
// (policy/binding). It will be merged with the base policy.
func radarResourceIAMSchema(binding bool) schema.Schema {
	d := "Sets the Vault Radar Resource IAM policy and replaces any existing policy."
	if binding {
		d = "Updates the Vault Radar Resource IAM policy to bind a role to a new principal. Existing bindings are preserved."
	}

	// Defined in: https://github.com/CyberAP/nanoid-dictionary#nolookalikessafe
	nanoidNolookalikesSafeAlphabet := "6789BCDFGHJKLMNPQRTWbcdfghjkmnpqrtwz"
	nanoidLength := 20
	resourceNameRegex := fmt.Sprintf(`^vault-radar/project/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}/scan-target/[%s]{%d}$`, nanoidNolookalikesSafeAlphabet, nanoidLength)

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				Required:    true,
				Description: "The HCP resource name associated with the Radar resource. This is the name of the resource in the format `vault-radar/project/<project_id>/scan-target/<scan_target_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(resourceNameRegex),
						"must match the format: "+resourceNameRegex,
					),
				},
			},
		},
	}
}

func NewRadarResourceIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("vault_radar_resource", radarResourceIAMSchema(false), "", newRadarResourceIAMPolicyUpdater)
}

func NewRadarResourceIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("vault_radar_resource", radarResourceIAMSchema(true), "", newRadarResourceIAMPolicyUpdater)
}

type radarResourceIAMPolicyUpdater struct {
	resourceName string
	client       *clients.Client
	d            iampolicy.TerraformResourceData
}

func newRadarResourceIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	var resourceName types.String
	diags := d.GetAttribute(ctx, path.Root("resource_name"), &resourceName)

	return &radarResourceIAMPolicyUpdater{
		resourceName: resourceName.ValueString(),
		client:       clients,
		d:            d,
	}, diags
}

func (u *radarResourceIAMPolicyUpdater) GetMutexKey() string {
	return u.resourceName
}

// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
func (u *radarResourceIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
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
func (u *radarResourceIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	allowableRolesSet := map[string]struct{}{
		"roles/vault-radar.resource-viewer":      {},
		"roles/vault-radar.resource-contributor": {},
	}
	allowableRolesMsg := strings.Join(slices.Collect(maps.Keys(allowableRolesSet)), ", ")

	for _, binding := range policy.Bindings {
		if _, ok := allowableRolesSet[binding.RoleID]; !ok {
			diags.AddError("invalid role in policy binding: "+binding.RoleID, fmt.Sprintf("allowable roles are: [%s]", allowableRolesMsg))
		}
	}

	if diags.HasError() {
		return nil, diags
	}

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
	_ iampolicy.NewResourceIamUpdaterFunc = newRadarResourceIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &radarResourceIAMPolicyUpdater{}
)
