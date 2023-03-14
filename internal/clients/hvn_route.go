// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// CreateHVNRoute creates a new HVN route
func CreateHVNRoute(ctx context.Context, client *Client,
	id string,
	hvn *sharedmodels.HashicorpCloudLocationLink,
	destination string,
	target *sharedmodels.HashicorpCloudLocationLink,
	location *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907CreateHVNRouteResponse, error) {

	hvnRouteParams := network_service.NewCreateHVNRouteParams()
	hvnRouteParams.Context = ctx
	hvnRouteParams.HvnLocationOrganizationID = location.OrganizationID
	hvnRouteParams.HvnLocationProjectID = location.ProjectID
	hvnRouteParams.HvnID = hvn.ID
	hvnRouteParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateHVNRouteRequest{
		Destination: destination,
		Hvn:         hvn,
		ID:          id,
		Target: &networkmodels.HashicorpCloudNetwork20200907HVNRouteTarget{
			HvnConnection: target,
		},
	}
	log.Printf("[INFO] Creating HVN route for HVN (%s) with destination CIDR %s", hvn.ID, destination)
	hvnRouteResp, err := client.Network.CreateHVNRoute(hvnRouteParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create HVN route for HVN (%s) with destination CIDR %s: %v", hvn.ID, destination, err)
	}

	return hvnRouteResp.Payload, nil
}

// GetHVNRoute returns specific HVN route by its ID
func GetHVNRoute(ctx context.Context, client *Client, hvnID, routeID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907HVNRoute, error) {
	getHVNRouteParams := network_service.NewGetHVNRouteParams()
	getHVNRouteParams.Context = ctx
	getHVNRouteParams.HvnID = hvnID
	getHVNRouteParams.ID = routeID
	getHVNRouteParams.HvnLocationOrganizationID = loc.OrganizationID
	getHVNRouteParams.HvnLocationProjectID = loc.ProjectID

	getHVNRouteResponse, err := client.Network.GetHVNRoute(getHVNRouteParams, nil)
	if err != nil {
		return nil, err
	}

	return getHVNRouteResponse.Payload.Route, nil
}

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

// hvnRouteRefreshState refreshes the state of the HVN route
func hvnRouteRefreshState(ctx context.Context, client *Client, hvnID, routeID string, loc *sharedmodels.HashicorpCloudLocationLocation) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		route, err := GetHVNRoute(ctx, client, hvnID, routeID, loc)
		if err != nil {
			return nil, "", err
		}

		return route, string(*route.State), nil
	}
}

// WaitForHVNRouteToBeActive will poll the GET HVN route endpoint until
// the state is ACTIVE, ctx is canceled, or an error occurs.
func WaitForHVNRouteToBeActive(ctx context.Context, client *Client,
	hvnID string,
	routeID string,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907HVNRoute, error) {

	stateChangeConf := resource.StateChangeConf{
		Pending: []string{
			HvnRouteStateCreating,
			HvnRouteStatePending,
		},
		Target: []string{
			HvnRouteStateActive,
		},
		Refresh:      hvnRouteRefreshState(ctx, client, hvnID, routeID, loc),
		Timeout:      timeout,
		PollInterval: 5 * time.Second,
	}

	result, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for the HVN route (%s) to become 'ACTIVE': %+v", routeID, err)
	}

	return result.(*networkmodels.HashicorpCloudNetwork20200907HVNRoute), nil
}
