// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	"github.com/cenkalti/backoff/v4"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	resourcemodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
)

// GetProjectByID gets a project by its ID
func GetProjectByID(ctx context.Context, client *Client, projectID string) (*resourcemodels.HashicorpCloudResourcemanagerProject, error) {
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

func CreateProject(ctx context.Context, client *Client, name, organizationID string) (*resourcemodels.HashicorpCloudResourcemanagerProject, error) {
	projectOrg := &resourcemodels.HashicorpCloudResourcemanagerResourceID{
		ID:   organizationID,
		Type: resourcemodels.NewHashicorpCloudResourcemanagerResourceIDResourceType("ORGANIZATION"),
	}
	projectParams := project_service.NewProjectServiceCreateParamsWithContext(ctx)
	projectParams.Body = &resourcemodels.HashicorpCloudResourcemanagerProjectCreateRequest{
		Name:   name,
		Parent: projectOrg,
	}

	createProjectResp, err := client.Project.ProjectServiceCreate(projectParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create project '%s' with organization ID %s: %v", name, organizationID, err)
	}

	return createProjectResp.Payload.Project, nil
}

// CreateProjectWithRetry wraps the projects service client with an exponential backoff retry mechanism.
func CreateProjectWithRetry(client *Client, params *project_service.ProjectServiceCreateParams) (*project_service.ProjectServiceCreateOK, error) {
	var res *project_service.ProjectServiceCreateOK
	op := func() error {
		var err error
		res, err = client.Project.ProjectServiceCreate(params, nil)
		// Wait for the project to be created, if an operation ID is returned.
		if res.Payload.OperationID != "" {
			return waitForProjectOperation(params.Context, client, "create project", res.Payload.Project.ID, res.Payload.OperationID)
		}
		return err
	}

	serviceErr := &project_service.ProjectServiceCreateDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// SetProjectNameWithRetry wraps the projects service client with an exponential backoff retry mechanism.
func SetProjectNameWithRetry(client *Client, params *project_service.ProjectServiceSetNameParams) (*project_service.ProjectServiceSetNameOK, error) {
	var res *project_service.ProjectServiceSetNameOK
	op := func() error {
		var err error
		res, err = client.Project.ProjectServiceSetName(params, nil)
		// Wait for the project name to be set, if an operation ID is returned.
		if res.Payload.OperationID != "" {
			return waitForProjectOperation(params.Context, client, "set project name", params.ID, res.Payload.OperationID)
		}
		return err
	}

	serviceErr := &project_service.ProjectServiceSetNameDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// SetProjectDescriptionWithRetry wraps the projects service client with an exponential backoff retry mechanism.
func SetProjectDescriptionWithRetry(client *Client, params *project_service.ProjectServiceSetDescriptionParams) (*project_service.ProjectServiceSetDescriptionOK, error) {
	var res *project_service.ProjectServiceSetDescriptionOK
	op := func() error {
		var err error
		res, err = client.Project.ProjectServiceSetDescription(params, nil)
		// Wait for the project description to be set, if an operation ID is returned.
		if res.Payload.OperationID != "" {
			return waitForProjectOperation(params.Context, client, "set project description", params.ID, res.Payload.OperationID)
		}
		return err
	}

	serviceErr := &project_service.ProjectServiceSetDescriptionDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// DeleteProjectWithRetry wraps the projects service client with an exponential backoff retry mechanism.
func DeleteProjectWithRetry(client *Client, params *project_service.ProjectServiceDeleteParams) (*project_service.ProjectServiceDeleteOK, error) {
	var res *project_service.ProjectServiceDeleteOK
	op := func() error {
		var err error
		res, err = client.Project.ProjectServiceDelete(params, nil)
		// Wait for the project to be deleted, if an operation ID is returned.
		if res.Payload.Operation.ID != "" {
			// For delete operations, the operation is scoped at the organization level
			projectID := ""
			return waitForProjectOperation(params.Context, client, "delete project", projectID, res.Payload.Operation.ID)
		}
		return err
	}

	serviceErr := &project_service.ProjectServiceDeleteDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

func waitForProjectOperation(ctx context.Context, client *Client, operationName, projectID string, operationID string) error {
	loc := &sharedmodels.HashicorpCloudLocationLocation{OrganizationID: client.Config.OrganizationID, ProjectID: projectID}
	return WaitForOperation(ctx, client, operationName, loc, operationID)
}
