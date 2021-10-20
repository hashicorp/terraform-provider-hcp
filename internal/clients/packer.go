package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetPackerChannelBySlug queries the HCP Packer registry for the iteration
// associated with the given channel name.
func GetPackerChannelBySlug(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channel string) (*packermodels.HashicorpCloudPackerChannel, error) {

	getParams := packer_service.NewPackerServiceGetChannelParams()
	getParams.BucketSlug = bucketName
	getParams.Slug = channel
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Packer.PackerServiceGetChannel(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Channel, nil
}

// GetIteration queries the HCP Packer registry for an existing bucket iteration.
func GetIterationFromId(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketslug string, iterationId string) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := packer_service.NewPackerServiceGetIterationParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketslug

	// The identifier can be either fingerprint, iterationid, or incremental version
	// for now, we only care about id so we're hardcoding it.
	params.IterationID = &iterationId

	it, err := client.Packer.PackerServiceGetIteration(params, nil)
	if err != nil {
		return nil, err
	}

	return it.Payload.Iteration, nil
}
