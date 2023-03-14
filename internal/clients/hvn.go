// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetHvnByID gets an HVN by its ID and location
func GetHvnByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, hvnID string) (*networkmodels.HashicorpCloudNetwork20200907Network, error) {
	getParams := network_service.NewGetParams()
	getParams.Context = ctx
	getParams.ID = hvnID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID
	getResponse, err := client.Network.Get(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResponse.Payload.Network, nil
}
