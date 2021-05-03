package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// ListHVNRoutes lists the routes for an HVN.
func ListHVNRoutes(ctx context.Context, client *Client, hvnID string,
	destination string, targetID string, targetType string,
	loc *sharedmodels.HashicorpCloudLocationLocation) ([]*networkmodels.HashicorpCloudNetwork20200907HVNRoute, error) {
	listHVNRoutesParams := network_service.NewListHVNRoutesParams()
	listHVNRoutesParams.Context = ctx
	listHVNRoutesParams.HvnID = hvnID
	listHVNRoutesParams.HvnLocationOrganizationID = loc.OrganizationID
	listHVNRoutesParams.HvnLocationProjectID = loc.ProjectID
	listHVNRoutesParams.Destination = &destination
	listHVNRoutesParams.TargetID = &targetID
	listHVNRoutesParams.TargetType = &targetType

	listHVNRoutesResponse, err := client.Network.ListHVNRoutes(listHVNRoutesParams, nil)
	if err != nil {
		return nil, err
	}

	return listHVNRoutesResponse.Payload.Routes, nil
}
