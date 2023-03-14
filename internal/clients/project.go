// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	resourcemodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
)

// GetProjectByID gets a project by its ID
func GetProjectByID(ctx context.Context, client *Client, projectID string) (*resourcemodels.ResourcemanagerProject, error) {
	getParams := project_service.NewProjectServiceGetParams()
	getParams.Context = ctx
	getParams.ID = projectID
	getResponse, err := client.Project.ProjectServiceGet(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResponse.Payload.Project, nil
}

// GetParentOrganizationIDByProjectID gets the parent organization ID of a project
func GetParentOrganizationIDByProjectID(ctx context.Context, client *Client, projectID string) (string, error) {
	project, err := GetProjectByID(ctx, client, projectID)
	if err != nil {
		return "", err
	}

	return project.Parent.ID, nil
}
