package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

// DeleteSnapshotByID deletes an HVN route by its ID
func DeleteHVNRouteByID(ctx context.Context, client *Client, hvnID string,
	hvnRouteID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907DeleteHVNRouteResponse, error) {
	deleteHVNRouteParams := network_service.NewDeleteHVNRouteParams()

	deleteHVNRouteParams.Context = ctx
	deleteHVNRouteParams.ID = hvnRouteID
	deleteHVNRouteParams.HvnID = hvnID
	deleteHVNRouteParams.HvnLocationOrganizationID = loc.OrganizationID
	deleteHVNRouteParams.HvnLocationProjectID = loc.ProjectID

	deleteHVNRouteResponse, err := client.Network.DeleteHVNRoute(deleteHVNRouteParams, nil)
	if err != nil {
		return nil, err
	}

	return deleteHVNRouteResponse.Payload, nil
}

const (
	// HvnRouteStateCreating is the CREATING state of an HVN route
	HvnRouteStateCreating = string(networkmodels.HashicorpCloudNetwork20200907HVNRouteStateCREATING)

	// HvnRouteStateActive is the ACTIVE state of an HVN route
	HvnRouteStateActive = string(networkmodels.HashicorpCloudNetwork20200907HVNRouteStateACTIVE)

	// HvnRouteStatePending is the PENDING state of an HVN route
	HvnRouteStatePending = string(networkmodels.HashicorpCloudNetwork20200907HVNRouteStatePENDING)
)

// hvnRouteRefreshState refreshes the state of the HVN route by calling
// the LIST endpoint
func hvnRouteRefreshState(ctx context.Context, client *Client, destination string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		route, err := ListHVNRoutes(ctx, client, hvnID, destination, "", "", loc)
		if err != nil {
			return nil, "", err
		}

		fmt.Println("here in refresh state")

		if len(route) != 1 {
			return nil, "", fmt.Errorf("Unexpected number of HVN route returned when waiting for route with destination CIDR of %s for HVN (%s) to be Active", destination, hvnID)
		}
		return route[0], string(route[0].State), nil
	}
}

// WaitForHVNRouteToBeActive will poll the LIST HVN route endpoint until
// the state is ACTIVE, ctx is canceled, or an error occurs.
func WaitForHVNRouteToBeActive(ctx context.Context, client *Client, destination string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907HVNRoute, error) {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{
			HvnRouteStateCreating,
			HvnRouteStatePending,
		},
		Target: []string{
			HvnRouteStateActive,
		},
		Refresh:      hvnRouteRefreshState(ctx, client, destination, hvnID, loc),
		Timeout:      timeout,
		PollInterval: 5 * time.Second,
	}

	result, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error waiting for route with destination CIDR of %s for HVN %s to become 'ACTIVE': %+v", destination, hvnID, err)
	}

	return result.(*networkmodels.HashicorpCloudNetwork20200907HVNRoute), nil
}

