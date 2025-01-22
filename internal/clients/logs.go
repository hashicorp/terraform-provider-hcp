// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/streaming_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/models"
)

// GetLogStreamingDestination will get an HCP Log Streaming Destination by its ID and location.
func GetLogStreamingDestination(ctx context.Context, client *Client, orgID, streamingDestinationID string) (*models.LogService20210330StreamingDestination, error) {
	getParams := streaming_service.NewStreamingServiceGetDestinationParams()
	getParams.Context = ctx
	getParams.DestinationID = streamingDestinationID
	getParams.OrganizationID = orgID

	getResponse, err := client.LogStreamingService.StreamingServiceGetDestination(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResponse.GetPayload().Destination, nil
}

// DeleteLogStreamingDestination will delete an HCP Log Streaming Destination.
func DeleteLogStreamingDestination(ctx context.Context, client *Client, orgID, streamingDestinationID string) error {
	deleteParams := streaming_service.NewStreamingServiceDeleteDestinationParams()
	deleteParams.Context = ctx
	deleteParams.DestinationID = streamingDestinationID
	deleteParams.OrganizationID = orgID

	_, err := client.LogStreamingService.StreamingServiceDeleteDestination(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLogStreamingDestination(ctx context.Context, client *Client, updatePaths []string, destination *models.LogService20210330StreamingDestination) error {
	updateParams := streaming_service.NewStreamingServiceUpdateDestinationParams()
	updateParams.Context = ctx
	updateParams.OrganizationID = destination.OrganizationID
	updateParams.ID = destination.ID

	updateBody := &models.LogService20210330UpdateDestinationRequest{
		OrganizationID:         destination.OrganizationID,
		ID:                     destination.ID,
		Name:                   destination.Name,
		DatadogProvider:        destination.DatadogProvider,
		CloudwatchLogsProvider: destination.CloudwatchLogsProvider,
		SplunkCloudProvider:    destination.SplunkCloudProvider,
		Mask: &models.ProtobufFieldMask{
			Paths: updatePaths,
		},
	}

	updateParams.Body = updateBody
	_, err := client.LogStreamingService.StreamingServiceUpdateDestination(updateParams, nil)
	if err != nil {
		return err
	}

	return nil
}
