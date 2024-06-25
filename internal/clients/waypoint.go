// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
)

// getNamespaceByLocation will retrieve a namespace by location information
// provided by HCP
func getNamespaceByLocation(_ context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation) (*waypoint_models.HashicorpCloudWaypointNamespace, error) {
	namespaceParams := &waypoint_service.WaypointServiceGetNamespaceParams{
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
func GetAction(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, actionID string, actionName string) (*waypoint_models.HashicorpCloudWaypointActionConfig, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetActionConfigParams{
		ActionID:    &actionID,
		ActionName:  &actionName,
		NamespaceID: ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetActionConfig(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ActionConfig, nil
}

// GetApplicationTemplateByName will retrieve a template by name
func GetApplicationTemplateByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*waypoint_models.HashicorpCloudWaypointApplicationTemplate, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetApplicationTemplate2Params{
		ApplicationTemplateName: appName,
		NamespaceID:             ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplicationTemplate2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ApplicationTemplate, nil
}

// GetApplicationTemplateByID will retrieve a template by ID
func GetApplicationTemplateByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appID string) (*waypoint_models.HashicorpCloudWaypointApplicationTemplate, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetApplicationTemplateParams{
		ApplicationTemplateID: appID,
		NamespaceID:           ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplicationTemplate(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().ApplicationTemplate, nil
}

// GetAddOnDefinitionByName will retrieve an add-on definition by name
func GetAddOnDefinitionByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defName string) (*waypoint_models.HashicorpCloudWaypointAddOnDefinition, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetAddOnDefinition2Params{
		AddOnDefinitionName: defName,
		NamespaceID:         ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOnDefinition2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOnDefinition, nil
}

// GetAddOnDefinitionByID will retrieve an add-on definition by ID
func GetAddOnDefinitionByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defID string) (*waypoint_models.HashicorpCloudWaypointAddOnDefinition, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetAddOnDefinitionParams{
		AddOnDefinitionID: defID,
		NamespaceID:       ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOnDefinition(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOnDefinition, nil
}

// GetApplicationByName will retrieve an application by name
func GetApplicationByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*waypoint_models.HashicorpCloudWaypointApplication, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetApplication2Params{
		ApplicationName: appName,
		NamespaceID:     ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplication2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().Application, nil
}

// GetApplicationByID will retrieve an application by ID
func GetApplicationByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appID string) (*waypoint_models.HashicorpCloudWaypointApplication, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetApplicationParams{
		ApplicationID: appID,
		NamespaceID:   ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetApplication(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().Application, nil
}

// GetAddOnByName will retrieve an add-on by name
func GetAddOnByName(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defName string) (*waypoint_models.HashicorpCloudWaypointAddOn, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetAddOn2Params{
		AddOnName:   defName,
		NamespaceID: ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOn2(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOn, nil
}

// GetAddOnByID will retrieve an add-on by ID
func GetAddOnByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, defID string) (*waypoint_models.HashicorpCloudWaypointAddOn, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetAddOnParams{
		AddOnID:     defID,
		NamespaceID: ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetAddOn(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().AddOn, nil
}

func GetInputVariables(ctx context.Context, client *Client, workspaceName string, loc *sharedmodels.HashicorpCloudLocationLocation) ([]*waypoint_models.HashicorpCloudWaypointInputVariable, error) {
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		return nil, err
	}

	params := &waypoint_service.WaypointServiceGetTFRunStatusParams{
		WorkspaceName: workspaceName,
		NamespaceID:   ns.ID,
	}

	getResp, err := client.Waypoint.WaypointServiceGetTFRunStatus(params, nil)
	if err != nil {
		return nil, err
	}
	return getResp.GetPayload().InputVariables, nil
}
