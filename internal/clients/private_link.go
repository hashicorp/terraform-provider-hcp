// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// GetPrivateLinkServiceByID gets a private link service by its ID, hvnID, and location
func GetPrivateLinkServiceByID(ctx context.Context, client *Client, privateLinkID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService, error) {
	getPrivateLinkParams := network_service.NewGetPrivateLinkServiceParams()
	getPrivateLinkParams.Context = ctx
	getPrivateLinkParams.ID = privateLinkID
	getPrivateLinkParams.HvnID = hvnID
	getPrivateLinkParams.HvnLocationOrganizationID = loc.OrganizationID
	getPrivateLinkParams.HvnLocationProjectID = loc.ProjectID
	getPrivateLinkResponse, err := client.Network.GetPrivateLinkService(getPrivateLinkParams, nil)
	if err != nil {
		return nil, err
	}

	return getPrivateLinkResponse.Payload.PrivateLinkService, nil
}

// ListPrivateLinkServices lists all private link services for an HVN
func ListPrivateLinkServices(ctx context.Context, client *Client, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) ([]*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService, error) {
	listPrivateLinkParams := network_service.NewListPrivateLinkServiceParams()
	listPrivateLinkParams.Context = ctx
	listPrivateLinkParams.HvnID = hvnID
	listPrivateLinkParams.HvnLocationOrganizationID = loc.OrganizationID
	listPrivateLinkParams.HvnLocationProjectID = loc.ProjectID

	listPrivateLinkResponse, err := client.Network.ListPrivateLinkService(listPrivateLinkParams, nil)
	if err != nil {
		return nil, err
	}

	return listPrivateLinkResponse.Payload.PrivateLinkServices, nil
}

// CreatePrivateLinkService creates a new private link service
func CreatePrivateLinkService(ctx context.Context, client *Client, privateLinkService *networkmodels.HashicorpCloudNetwork20200907PrivateLinkService) (*networkmodels.HashicorpCloudNetwork20200907CreatePrivateLinkServiceResponse, error) {
	createPrivateLinkParams := network_service.NewCreatePrivateLinkServiceParams()
	createPrivateLinkParams.Context = ctx
	createPrivateLinkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreatePrivateLinkServiceRequest{
		PrivateLinkService: privateLinkService,
	}
	createPrivateLinkParams.PrivateLinkServiceHvnID = privateLinkService.Hvn.ID
	createPrivateLinkParams.PrivateLinkServiceHvnLocationOrganizationID = privateLinkService.Hvn.Location.OrganizationID
	createPrivateLinkParams.PrivateLinkServiceHvnLocationProjectID = privateLinkService.Hvn.Location.ProjectID

	createPrivateLinkResponse, err := client.Network.CreatePrivateLinkService(createPrivateLinkParams, nil)
	if err != nil {
		return nil, err
	}

	return createPrivateLinkResponse.Payload, nil
}

// DeletePrivateLinkServiceByID deletes a private link service by its ID, hvnID, and location
func DeletePrivateLinkServiceByID(ctx context.Context, client *Client, privateLinkID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907DeletePrivateLinkServiceResponse, error) {
	deletePrivateLinkParams := network_service.NewDeletePrivateLinkServiceParams()
	deletePrivateLinkParams.Context = ctx
	deletePrivateLinkParams.ID = privateLinkID
	deletePrivateLinkParams.HvnID = hvnID
	deletePrivateLinkParams.LocationOrganizationID = loc.OrganizationID
	deletePrivateLinkParams.LocationProjectID = loc.ProjectID

	deletePrivateLinkResponse, err := client.Network.DeletePrivateLinkService(deletePrivateLinkParams, nil)
	if err != nil {
		return nil, err
	}

	return deletePrivateLinkResponse.Payload, nil
}

// UpdatePrivateLinkService updates a private link service
func UpdatePrivateLinkService(
	ctx context.Context,
	client *Client,
	privateLinkID string,
	hvnID string,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	addConsumerRegions []string,
	removeConsumerRegions []string,
	addConsumerAccounts []string,
	removeConsumerAccounts []string,
	addConsumerIPRanges []string,
	removeConsumerIPRanges []string,
) (*networkmodels.HashicorpCloudNetwork20200907UpdatePrivateLinkServiceResponse, error) {
	updatePrivateLinkParams := network_service.NewUpdatePrivateLinkServiceParams()
	updatePrivateLinkParams.Context = ctx
	updatePrivateLinkParams.ID = privateLinkID
	updatePrivateLinkParams.HvnID = hvnID
	updatePrivateLinkParams.HvnLocationOrganizationID = loc.OrganizationID
	updatePrivateLinkParams.HvnLocationProjectID = loc.ProjectID
	updatePrivateLinkParams.Body = &networkmodels.HashicorpCloudNetwork20200907UpdatePrivateLinkServiceRequest{
		AddConsumerRegions:     addConsumerRegions,
		RemoveConsumerRegions:  removeConsumerRegions,
		AddConsumerAccounts:    addConsumerAccounts,
		RemoveConsumerAccounts: removeConsumerAccounts,
		AddConsumerIPRanges:    addConsumerIPRanges,
		RemoveConsumerIPRanges: removeConsumerIPRanges,
	}

	updatePrivateLinkResponse, err := client.Network.UpdatePrivateLinkService(updatePrivateLinkParams, nil)
	if err != nil {
		return nil, err
	}

	return updatePrivateLinkResponse.Payload, nil
}

const (
	// PrivateLinkServiceStateCreating is the CREATING state of a private link service
	PrivateLinkServiceStateCreating = string(networkmodels.HashicorpCloudNetwork20200907PrivateLinkServiceStateCREATING)

	// PrivateLinkServiceStateAvailable is the AVAILABLE state of a private link service
	PrivateLinkServiceStateAvailable = string(networkmodels.HashicorpCloudNetwork20200907PrivateLinkServiceStateAVAILABLE)

	// PrivateLinkServiceStateDeleting is the DELETING state of a private link service
	PrivateLinkServiceStateDeleting = string(networkmodels.HashicorpCloudNetwork20200907PrivateLinkServiceStateDELETING)

	// PrivateLinkServiceStateFailed is the FAILED state of a private link service
	PrivateLinkServiceStateFailed = string(networkmodels.HashicorpCloudNetwork20200907PrivateLinkServiceStateFAILED)

	// PrivateLinkServiceStateUpdating is the UPDATING state of a private link service
	PrivateLinkServiceStateUpdating = string(networkmodels.HashicorpCloudNetwork20200907PrivateLinkServiceStateUPDATING)
)

// privateLinkServiceRefreshState refreshes the state of the private link service by calling
// the GET endpoint
func privateLinkServiceRefreshState(ctx context.Context, client *Client, privateLinkID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		privateLinkService, err := GetPrivateLinkServiceByID(ctx, client, privateLinkID, hvnID, loc)
		if err != nil {
			return nil, "", err
		}

		return privateLinkService, string(*privateLinkService.State), nil
	}
}

type PrivateLinkServiceWaitFor = func(ctx context.Context, client *Client, privateLinkID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService, error)

// privateLinkServiceState contains a target private link service state and a list of every allowed pending state
type privateLinkServiceState struct {
	Target  string
	Pending []string
}

func waitForPrivateLinkServiceToBe(ps privateLinkServiceState) PrivateLinkServiceWaitFor {
	return func(ctx context.Context, client *Client, privateLinkID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService, error) {
		stateChangeConfig := retry.StateChangeConf{
			Pending: ps.Pending,
			Target: []string{
				ps.Target,
			},
			Refresh:      privateLinkServiceRefreshState(ctx, client, privateLinkID, hvnID, loc),
			Timeout:      timeout,
			PollInterval: 5 * time.Second,
		}

		result, err := stateChangeConfig.WaitForStateContext(ctx)
		if err != nil {
			err = fmt.Errorf("error waiting for private link service (%s) to become '%s': %v", privateLinkID, ps.Target, err)
			if result != nil {
				return result.(*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService), err
			}
			return nil, err
		}

		return result.(*networkmodels.HashicorpCloudNetwork20200907PrivateLinkService), nil
	}
}

// WaitForPrivateLinkServiceToBeAvailable will poll the GET private link service endpoint until
// the state is AVAILABLE, ctx is canceled, or an error occurs.
var WaitForPrivateLinkServiceToBeAvailable = waitForPrivateLinkServiceToBe(privateLinkServiceState{
	Target:  PrivateLinkServiceStateAvailable,
	Pending: []string{PrivateLinkServiceStateCreating, PrivateLinkServiceStateUpdating},
})

// WaitForPrivateLinkServiceToBeFailed will poll the GET private link service endpoint until
// the state is FAILED, ctx is canceled, or an error occurs.
var WaitForPrivateLinkServiceToBeFailed = waitForPrivateLinkServiceToBe(privateLinkServiceState{
	Target:  PrivateLinkServiceStateFailed,
	Pending: []string{PrivateLinkServiceStateCreating, PrivateLinkServiceStateUpdating},
})
