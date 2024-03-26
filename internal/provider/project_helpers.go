// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	diagnostic "github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// getProjectFromCredentials uses the configured client credentials to
// fetch the associated organization and returns that organization's
// single project.
// This differs from the provider.go implementation due to the diagnostics used
// by the plugin framework.
func getProjectFromCredentialsFramework(ctx context.Context, client *clients.Client) (project *models.HashicorpCloudResourcemanagerProject, diags diagnostic.Diagnostics) {
	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := clients.RetryOrganizationServiceList(client, listOrgParams)
	if err != nil {
		diags.AddError(fmt.Sprintf("unable to fetch organization list: %v", err), "")

		return nil, diags
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen == 0 {
		diags.AddError("The configured credentials do not have access to any organization.", "Please assign at least one organization to the configured credentials to use this provider.")
		return nil, diags
	}
	if orgLen > 1 {
		diags.AddError("There is more than one organization associated with the configured credentials.", "Please configure a specific project in the HCP provider config block.")
		return nil, diags
	}

	orgID := listOrgResp.Payload.Organizations[0].ID

	// Get the project using the organization ID.
	listProjParams := project_service.NewProjectServiceListParams()
	listProjParams.ScopeID = &orgID
	scopeType := string(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := clients.RetryProjectServiceList(client, listProjParams)
	if err != nil {
		diags.AddError(fmt.Sprintf("unable to fetch project id: %v", err), "")
		return nil, diags
	}
	if len(listProjResp.Payload.Projects) == 0 {
		diags.AddError("The configured credentials does not have access to any project.", "Please assign at least one project to the configured credentials to use this provider.")
		return nil, diags
	}
	if len(listProjResp.Payload.Projects) > 1 {
		diags.AddWarning("There is more than one project associated with the organization of the configured credentials.", `The oldest project has been selected as the default. To configure which project is used as default, set a project in the HCP provider config block. Resources may also be configured with different projects.`)
		return getOldestProject(listProjResp.Payload.Projects), diags
	}
	project = listProjResp.Payload.Projects[0]
	return project, diags
}

// getOldestProject retrieves the oldest project from a list based on its created_at time.
func getOldestProject(projects []*models.HashicorpCloudResourcemanagerProject) (oldestProj *models.HashicorpCloudResourcemanagerProject) {
	oldestTime := time.Now()

	for _, proj := range projects {
		projTime := time.Time(proj.CreatedAt)
		if projTime.Before(oldestTime) {
			oldestProj = proj
			oldestTime = projTime
		}
	}
	return oldestProj
}
