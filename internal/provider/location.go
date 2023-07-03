// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/go-uuid"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func GetProjectID(resourceProjID, clientProjID string) (string, error) {
	if resourceProjID != "" {
		return resourceProjID, nil
	} else {
		if clientProjID != "" {
			return clientProjID, nil
		} else {
			return "", fmt.Errorf("project ID not defined. Verify that project ID is set either in the provider or in the resource config")
		}
	}
}

func setLocationResourceData(d *schema.ResourceData, loc *sharedmodels.HashicorpCloudLocationLocation) error {
	if loc == nil {
		return fmt.Errorf("failed to set location attributes, expected non-nil location, got nil")
	}
	if _, err := uuid.ParseUUID(loc.OrganizationID); err != nil {
		return fmt.Errorf("expected Organization ID to be a valid UUID, got %q", loc.OrganizationID)
	}
	if _, err := uuid.ParseUUID(loc.ProjectID); err != nil {
		return fmt.Errorf("expected Project ID to be a valid UUID, got %q", loc.ProjectID)
	}

	if err := d.Set("organization_id", loc.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}

	return nil
}

func getLocationResourceData(d *schema.ResourceData, client *clients.Client) (*sharedmodels.HashicorpCloudLocationLocation, error) {
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return nil, err
	}
	return &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}, nil
}

func getAndUpdateLocationResourceData(d *schema.ResourceData, client *clients.Client) (*sharedmodels.HashicorpCloudLocationLocation, error) {
	loc, err := getLocationResourceData(d, client)
	if err != nil {
		return nil, fmt.Errorf("error while getting location: %v", err)
	}

	if err := setLocationResourceData(d, loc); err != nil {
		return nil, fmt.Errorf("error while setting location attriutes: %v", err)
	}

	return loc, nil
}
