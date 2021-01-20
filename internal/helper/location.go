package helper

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// BuildResourceLocation builds a Hashicorp Cloud Location based off of the resource data's
// org, project, region, and provider details
func BuildResourceLocation(ctx context.Context, d *schema.ResourceData, client *clients.Client, resType string) (
	*sharedmodels.HashicorpCloudLocationLocation, error) {

	provider := d.Get("cloud_provider").(string)
	region := d.Get("region").(string)

	projectID := client.Config.ProjectID
	projectIDVal, ok := d.GetOk("project_id")
	if ok {
		projectID = projectIDVal.(string)
	}

	organizationID := client.Config.OrganizationID
	organizationIDVal, ok := d.GetOk("organization_id")
	if ok {
		organizationID = organizationIDVal.(string)
	}

	if projectID == "" {
		return nil, fmt.Errorf("missing project_id: a project_id must be specified on the %s resource or the provider", resType)
	}

	organizationID, err := clients.GetParentOrganizationIDByProjectID(ctx, client, projectID)

	if organizationID == "" {
		if organizationID, err = clients.GetParentOrganizationIDByProjectID(ctx, client, projectID); err != nil {
			return nil, fmt.Errorf("unable to retrieve organization ID for project [project_id=%s]: %v", projectID, err)
		}
	}

	return &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: provider,
			Region:   region,
		},
	}, nil
}
