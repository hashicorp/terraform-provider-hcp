// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/log_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// CreateLogStreamingDestinationOrgFilter will create an HCP Log Streaming Destination Organization Filter.
func CreateLogStreamingDestinationOrgFilter(ctx context.Context, client *Client, orgID string, streamingDestinationID string) error {
	filter := &models.LogService20210330OrganizationFilter{}
	createParams := log_service.NewLogServiceCreateStreamingDestinationFilterParams()
	createParams.Context = ctx
	createParams.DestinationID = streamingDestinationID
	createParams.OrganizationID = orgID
	createParams.Body = &models.LogService20210330CreateStreamingDestinationFilterRequest{
		OrganizationFilter: filter,
		DestinationID:      streamingDestinationID,
		OrganizationID:     orgID,
	}

	_, err := client.LogService.LogServiceCreateStreamingDestinationFilter(createParams, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetLogStreamingDestination will get an HCP Log Streaming Destination by its ID and location.
func GetLogStreamingDestination(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, streamingDestinationID string) (*models.LogService20210330Destination, error) {
	getParams := log_service.NewLogServiceGetStreamingDestinationParams()
	getParams.Context = ctx
	getParams.DestinationID = streamingDestinationID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResponse, err := client.LogService.LogServiceGetStreamingDestination(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResponse.Payload.Destination, nil
}

// DeleteLogStreamingDestination will delete an HCP Log Streaming Destination.
func DeleteLogStreamingDestination(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, streamingDestinationID string) error {
	deleteParams := log_service.NewLogServiceDeleteStreamingDestinationParams()
	deleteParams.Context = ctx
	deleteParams.DestinationID = streamingDestinationID
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationProjectID = loc.ProjectID

	_, err := client.LogService.LogServiceDeleteStreamingDestination(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLogStreamingDestination(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, updatePaths []string, destination *models.LogService20210330Destination) error {
	updateParams := log_service.NewLogServiceUpdateStreamingDestinationParams()
	updateParams.Context = ctx
	updateParams.DestinationResourceID = destination.Resource.ID
	updateParams.DestinationResourceLocationOrganizationID = loc.OrganizationID
	updateParams.DestinationResourceLocationProjectID = loc.ProjectID

	updateBody := &models.LogService20210330UpdateStreamingDestinationRequest{
		Destination: destination,
		Mask: &models.ProtobufFieldMask{
			Paths: updatePaths,
		},
	}

	updateParams.Body = updateBody
	_, err := client.LogService.LogServiceUpdateStreamingDestination(updateParams, nil)
	if err != nil {
		return err
	}

	return nil
}
