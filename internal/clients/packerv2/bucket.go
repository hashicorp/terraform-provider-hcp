// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerv2

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type Bucket = packermodels.HashicorpCloudPacker20230101Bucket

// ListBuckets queries the HCP Packer registry for all associated buckets.
func ListBuckets(ctx context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation) ([]*Bucket, error) {
	nextPage := ""
	var buckets []*Bucket

	for {
		params := packerservice.NewPackerServiceListBucketsParams()
		params.LocationOrganizationID = loc.OrganizationID
		params.LocationProjectID = loc.ProjectID
		// Sort order is needed for acceptance tests.
		params.SortingOrderBy = []string{"name"}
		if nextPage != "" {
			params.PaginationNextPageToken = &nextPage
		}

		req, err := client.PackerV2.PackerServiceListBuckets(params, nil)
		if err != nil {
			if err, ok := err.(*packer_service.PackerServiceListBucketsDefault); ok {
				return nil, errors.New(err.Payload.Message)
			}
			return nil, fmt.Errorf("unexpected error format received by ListBuckets. Got: %v", err)
		}

		buckets = append(buckets, req.Payload.Buckets...)
		pagination := req.Payload.Pagination
		if pagination == nil || pagination.NextPageToken == "" {
			return buckets, nil
		}

		nextPage = pagination.NextPageToken
	}
}

func CreateBucket(ctx context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation, name string) (*Bucket, error) {
	params := packerservice.NewPackerServiceCreateBucketParams()
	params.SetLocationOrganizationID(loc.OrganizationID)
	params.SetLocationProjectID(loc.ProjectID)
	params.Body = &packermodels.HashicorpCloudPacker20230101CreateBucketBody{
		Name: name,
	}

	resp, err := client.PackerV2.PackerServiceCreateBucket(params, nil)

	if err != nil {
		return nil, formatGRPCError[*packerservice.PackerServiceGetBucketDefault](err)
	}
	return resp.GetPayload().Bucket, nil
}
