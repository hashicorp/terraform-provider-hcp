// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerv1

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// GetPackerChannelBySlug queries the HCP Packer registry for the channel
// associated with the given channel name.
func GetPackerChannelBySlug(ctx context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string) (*packermodels.HashicorpCloudPackerChannel, error) {

	getParams := packer_service.NewPackerServiceGetChannelParamsWithContext(ctx)
	getParams.BucketSlug = bucketName
	getParams.Slug = channelName
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Packer.PackerServiceGetChannel(getParams, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceGetChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by GetPackerChannelBySlug. Got: %v", err)
	}

	return getResp.Payload.Channel, nil
}

// GetIterationFromID queries the HCP Packer registry for an existing bucket iteration using its ULID.
func GetIterationFromID(ctx context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, iterationID string) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := newGetIterationParams(ctx, loc, bucketSlug)
	params.IterationID = &iterationID
	return getIteration(client, params)
}

// GetIterationFromFingerprint queries the HCP Packer registry for an existing bucket iteration using its fingerprint.
func GetIterationFromFingerprint(ctx context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, iterationFingerprint string) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := newGetIterationParams(ctx, loc, bucketSlug)
	params.Fingerprint = &iterationFingerprint
	return getIteration(client, params)
}

func newGetIterationParams(ctx context.Context, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketslug string) *packer_service.PackerServiceGetIterationParams {
	params := packer_service.NewPackerServiceGetIterationParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketslug
	return params
}

func getIteration(client *clients.Client, params *packer_service.PackerServiceGetIterationParams) (*packermodels.HashicorpCloudPackerIteration, error) {
	it, err := client.Packer.PackerServiceGetIteration(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceGetIterationDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by getIteration. Got: %v", err)
	}

	return it.Payload.Iteration, nil
}
