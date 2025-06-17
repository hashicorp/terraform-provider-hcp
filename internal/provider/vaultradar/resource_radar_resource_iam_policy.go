// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	rrs "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/resource_service"
	rrm "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/models"
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

	return schema.Schema{
		MarkdownDescription: d,
		Attributes: map[string]schema.Attribute{
			"resource_uri": schema.StringAttribute{
				Required:    true,
				Description: "The project's Radar resource URI.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
	resourceURI string
	client      *clients.Client
	d           iampolicy.TerraformResourceData
}

func newRadarResourceIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {

	var resourceURI types.String
	diags := d.GetAttribute(ctx, path.Root("resource_uri"), &resourceURI)

	return &radarResourceIAMPolicyUpdater{
		resourceURI: resourceURI.ValueString(),
		client:      clients,
		d:           d,
	}, diags
}

func (u *radarResourceIAMPolicyUpdater) GetMutexKey() string {
	return u.resourceURI
}

// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
func (u *radarResourceIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	rr, lookupDiags := lookupRadarResourceByURI(ctx, u.client, u.resourceURI)
	if lookupDiags.HasError() {
		return nil, lookupDiags
	}

	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceGetIamPolicyParams()
	params.ResourceName = &rr.HcpResourceName

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
	rr, lookupDiags := lookupRadarResourceByURI(ctx, u.client, u.resourceURI)
	if lookupDiags.HasError() {
		return nil, lookupDiags
	}

	var diags diag.Diagnostics
	params := resource_service.NewResourceServiceSetIamPolicyParams()

	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		Policy:       policy,
		ResourceName: rr.HcpResourceName,
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

func lookupRadarResourceByURI(ctx context.Context, client *clients.Client, resourceURI string) (*rrm.VaultRadar20230501Resource, diag.Diagnostics) {
	projectID := client.Config.ProjectID

	// Filter to find the exact radar resources by URI.
	filters := []*rrm.VaultRadar20230501Filter{
		{
			ID:         "uri",
			Op:         rrm.NewFilterFilterOperation(rrm.FilterFilterOperationEQ),
			Value:      []*rrm.VaultRadar20230501FilterValue{{StringValue: resourceURI}},
			ExactMatch: true,
		},
		{
			ID:         "state",
			Op:         rrm.NewFilterFilterOperation(rrm.FilterFilterOperationNEQNULLAWARE),
			Value:      []*rrm.VaultRadar20230501FilterValue{{StringValue: "deleted"}},
			ExactMatch: true,
		},
	}

	body := rrs.ListResourcesBody{
		Location: &rrs.ListResourcesParamsBodyLocation{
			OrganizationID: client.Config.OrganizationID,
		},
		Search: &rrm.VaultRadar20230501SearchSchema{
			Limit:   1, // we expect either 0 or 1
			Page:    1,
			Filters: filters,
		},
	}

	var diags diag.Diagnostics
	resp, err := clients.ListRadarResources(ctx, client, projectID, body)
	if err != nil {
		var srvErr *rrs.ListResourcesDefault
		ok := errors.As(err, &srvErr)
		if !ok {
			diags.AddError("unexpected error while reading response from service.", err.Error())
			return nil, diags
		}

		diags.Append(customdiags.NewErrorHTTPStatusCode("failed to retrieve radar resource", err.Error(), srvErr.Code()))
		return nil, diags
	}

	resources := resp.GetPayload().Resources
	if len(resources) == 0 {
		diags.AddError("unable to find radar resource", "no radar resource could be found with the uri: "+resourceURI)
		return nil, diags
	}

	resource := resources[0]
	diags.Append(validateResource(resource, projectID)...)

	return resources[0], diags
}

func validateResource(resource *rrm.VaultRadar20230501Resource, projectID string) diag.Diagnostics {
	var diags diag.Diagnostics

	if projectID == "" {
		diags.AddError("hcp project_id unknown", "the project_id where Vault Radar is located must be specified in the hcp provider config.")
		return diags
	}

	if resource.HcpResourceStatus != "registered" {
		diags.AddError("invalid radar resource status", "policy cannot be applied to a resource with status: "+resource.HcpResourceStatus)
	}

	if resource.State != "created" {
		diags.AddError("invalid radar resource state", "policy cannot be applied to a resource with state: "+resource.State)
	}

	// This should never happen, but we check it just in case.
	prefix := "vault-radar/project/" + projectID + "/scan-target/"
	if !strings.HasPrefix(resource.HcpResourceName, prefix) {
		diags.AddError("invalid radar resource name", "got resource name: "+resource.HcpResourceName+", expected it to start with: "+prefix)
	}

	return diags
}
