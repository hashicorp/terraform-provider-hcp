package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetPackerBucketByID gets an HCP bucket from its id
func GetPackerBucketByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bid string) (*packermodels.HashicorpCloudPackerBucket, error) {

	getParams := packer_service.NewGetBucketParams()
	getParams.Context = ctx
	getParams.BucketID = &bid
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Packer.GetBucket(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Bucket, nil
}

// GetPackerImageByChannelName
func GetPackerIterationByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucket *packermodels.HashicorpCloudPackerBucket,
	iid string) (*packermodels.HashicorpCloudPackerIteration, error) {

	getParams := packer_service.NewGetIterationParams()
	getParams.BucketSlug = &bucket.Slug
	getParams.IterationID = iid

	getResp, err := client.Packer.GetIteration(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Iteration, nil
}
