// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	waypoint_service_v2 "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	waypoint_models_v2 "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
)

// getNamespaceByLocation will retrieve a namespace by location information
// provided by HCP
func getNamespaceByLocation(_ context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation) (*waypoint_models_v2.HashicorpCloudWaypointNamespace, error) {
	namespaceParams := &waypoint_service_v2.WaypointServiceGetNamespaceParams{
		LocationOrganizationID: loc.OrganizationID,
		LocationProjectID:      loc.ProjectID,
	}
	// get namespace
	ns, err := client.Waypoint.WaypointServiceGetNamespace(namespaceParams, nil)
	if err != nil {
		return nil, err
	}
	return ns.GetPayload().Namespace, nil
}

// GetAction will retrieve an Action using the provided ID by default
// or by name if the ID is not provided
func GetAction(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, actionID string, actionName string) (*waypoint_models_v2.HashicorpCloudWaypointActionConfig, error) {
	params := &waypoint_service_v2.WaypointServiceGetActionConfigParams{
		ActionID:                        &actionID,
		ActionName:                      &actionName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetActionConfig(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ActionConfig, nil
}

// GetApplicationTemplateByName will retrieve a template by name
func GetApplicationTemplateByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*waypoint_models_v2.HashicorpCloudWaypointApplicationTemplate, error) {
	params := &waypoint_service_v2.WaypointServiceGetApplicationTemplate2Params{
		ApplicationTemplateName:         appName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplicationTemplate2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ApplicationTemplate, nil
}

// GetApplicationTemplateByID will retrieve a template by ID
func GetApplicationTemplateByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appID string) (*waypoint_models_v2.HashicorpCloudWaypointApplicationTemplate, error) {
	params := &waypoint_service_v2.WaypointServiceGetApplicationTemplateParams{
		ApplicationTemplateID:           appID,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplicationTemplate(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ApplicationTemplate, nil
}

// GetAddOnDefinitionByName will retrieve an add-on definition by name
func GetAddOnDefinitionByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defName string) (*waypoint_models_v2.HashicorpCloudWaypointAddOnDefinition, error) {
	params := &waypoint_service_v2.WaypointServiceGetAddOnDefinition2Params{
		AddOnDefinitionName:             defName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOnDefinition2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOnDefinition, nil
}

// GetAddOnDefinitionByID will retrieve an add-on definition by ID
func GetAddOnDefinitionByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defID string) (*waypoint_models_v2.HashicorpCloudWaypointAddOnDefinition, error) {
	params := &waypoint_service_v2.WaypointServiceGetAddOnDefinitionParams{
		AddOnDefinitionID:               defID,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOnDefinition(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOnDefinition, nil
}

// GetApplicationByName will retrieve an application by name
func GetApplicationByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*waypoint_models_v2.HashicorpCloudWaypointApplication, error) {
	params := &waypoint_service_v2.WaypointServiceGetApplication2Params{
		ApplicationName:                 appName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplication2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().Application, nil
}

// GetApplicationByID will retrieve an application by ID
func GetApplicationByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appID string) (*waypoint_models_v2.HashicorpCloudWaypointApplication, error) {
	params := &waypoint_service_v2.WaypointServiceGetApplicationParams{
		ApplicationID:                   appID,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplication(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().Application, nil
}

// GetAddOnByName will retrieve an add-on by name
func GetAddOnByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defName string) (*waypoint_models_v2.HashicorpCloudWaypointAddOn, error) {
	params := &waypoint_service_v2.WaypointServiceGetAddOn2Params{
		AddOnName:                       defName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOn2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOn, nil
}

// GetAddOnByID will retrieve an add-on by ID
func GetAddOnByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defID string) (*waypoint_models_v2.HashicorpCloudWaypointAddOn, error) {
	params := &waypoint_service_v2.WaypointServiceGetAddOnParams{
		AddOnID:                         defID,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOn(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOn, nil
}

func GetInputVariables(ctx context.Context, client *Client, workspaceName string, loc *sharedmodels.HashicorpCloudLocationLocation) ([]*waypoint_models_v2.HashicorpCloudWaypointInputVariable, error) {
	params := &waypoint_service_v2.WaypointServiceGetTFRunStatusParams{
		WorkspaceName:                   workspaceName,
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetTFRunStatus(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().InputVariables, nil
}
