// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerv2

import (
	"context"
	"errors"
	"fmt"
	"strings"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

type Channel = packermodels.HashicorpCloudPacker20230101Channel
type GetChannelParams = packerservice.PackerServiceGetChannelParams

func GetChannelByName(client *clients.Client, location location.BucketLocation, name string) (*Channel, error) {
	params := packerservice.NewPackerServiceGetChannelParams()
	params.SetLocationOrganizationID(location.GetOrganizationID())
	params.SetLocationProjectID(location.GetProjectID())
	params.SetBucketName(location.GetBucketName())
	params.SetChannelName(name)

	resp, err := getChannelRaw(client, params)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload().Channel, nil
}

func getChannelRaw(client *clients.Client, params *GetChannelParams) (*packerservice.PackerServiceGetChannelOK, error) {
	resp, err := client.PackerV2.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, formatGRPCError[*packerservice.PackerServiceGetChannelDefault](err)
	}

	return resp, nil
}

// CreatePackerChannel creates a channel in tge given bucket.
func CreatePackerChannel(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string, restricted bool,
) (*Channel, error) {
	params := packerservice.NewPackerServiceCreateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.BucketName = bucketName
	params.Body = &packermodels.HashicorpCloudPacker20230101CreateChannelBody{
		Name:       channelName,
		Restricted: restricted,
	}

	channel, err := client.PackerV2.PackerServiceCreateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceCreateChannelDefault); ok {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected error format received by PackerServiceCreateChannel. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}

// UpdatePackerChannel updates the named channel.
func UpdatePackerChannel(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string, restricted bool,
) (*Channel, error) {
	params := packerservice.NewPackerServiceUpdateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName

	params.Body = &packermodels.HashicorpCloudPacker20230101UpdateChannelBody{}
	params.Body.UpdateMask = "restricted"
	params.Body.Restricted = restricted

	channel, err := client.PackerV2.PackerServiceUpdateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceUpdateChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by UpdatePackerChannel. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}

// GetPackerChannelByNameFromList queries the HCP Packer Registry for the channel associated with the given
// channel name, using ListBucketChannels
func GetPackerChannelByNameFromList(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation, bucketName string,
	channelName string,
) (*Channel, error) {
	params := packerservice.NewPackerServiceListChannelsParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.BucketName = bucketName

	resp, err := client.PackerV2.PackerServiceListChannels(params, nil)
	if err != nil {
		return nil, err
	}

	for _, channel := range resp.GetPayload().Channels {
		if channel.Name == channelName {
			return channel, nil
		}
	}

	return nil, nil
}

// DeletePackerChannel deletes a channel from the named bucket.
func DeletePackerChannel(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
	bucketName, channelName string,
) (*Channel, error) {
	params := packerservice.NewPackerServiceDeleteChannelParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName

	req, err := client.PackerV2.PackerServiceDeleteChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceDeleteChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by DeletePackerChannel. Got: %v", err)
	}

	if !req.IsSuccess() {
		return nil, errors.New("failed to delete channel")
	}

	return nil, nil
}

func UpdatePackerChannelAssignment(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
	bucketName, channelName, versionFingerprint string,
) (*Channel, error) {
	params := packerservice.NewPackerServiceUpdateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName
	params.Body = &packermodels.HashicorpCloudPacker20230101UpdateChannelBody{}

	maskPaths := []string{}
	params.Body.VersionFingerprint = versionFingerprint
	maskPaths = append(maskPaths, "versionFingerprint")
	params.Body.UpdateMask = strings.Join(maskPaths, ",")

	channel, err := client.PackerV2.PackerServiceUpdateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceUpdateChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by UpdatePackerChannelAssignment. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}
