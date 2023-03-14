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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// GetPeeringByID gets a peering by its ID, hvnID, and location
func GetPeeringByID(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907Peering, error) {
	getPeeringParams := network_service.NewGetPeeringParams()
	getPeeringParams.Context = ctx
	getPeeringParams.ID = peeringID
	getPeeringParams.HvnID = hvnID
	getPeeringParams.LocationOrganizationID = loc.OrganizationID
	getPeeringParams.LocationProjectID = loc.ProjectID
	getPeeringResponse, err := client.Network.GetPeering(getPeeringParams, nil)
	if err != nil {
		return nil, err
	}

	return getPeeringResponse.Payload.Peering, nil
}

const (
	// PeeringStateCreating is the CREATING state of a peering connection
	PeeringStateCreating = string(networkmodels.HashicorpCloudNetwork20200907PeeringStateCREATING)

	// PeeringStatePendingAcceptance is the PENDING_ACCEPTANCE state of a peering connection
	PeeringStatePendingAcceptance = string(networkmodels.HashicorpCloudNetwork20200907PeeringStatePENDINGACCEPTANCE)

	// PeeringStateAccepted is the ACCEPTED state of a peering connection
	PeeringStateAccepted = string(networkmodels.HashicorpCloudNetwork20200907PeeringStateACCEPTED)

	// PeeringStateActive is the ACTIVE state of a peering connection
	PeeringStateActive = string(networkmodels.HashicorpCloudNetwork20200907PeeringStateACTIVE)
)

// peeringRefreshState refreshes the state of the peering connection by calling
// the GET endpoint
func peeringRefreshState(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		peering, err := GetPeeringByID(ctx, client, peeringID, hvnID, loc)
		if err != nil {
			return nil, "", err
		}

		return peering, string(*peering.State), nil
	}
}

type WaitFor = func(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907Peering, error)

// peeringState contains a target peering state and a list of every allowed pending state
type peeringState struct {
	Target  string
	Pending []string
}

func waitForPeeringToBe(ps peeringState) WaitFor {
	return func(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907Peering, error) {
		stateChangeConfig := resource.StateChangeConf{
			Pending: ps.Pending,
			Target: []string{
				ps.Target,
			},
			Refresh:      peeringRefreshState(ctx, client, peeringID, hvnID, loc),
			Timeout:      timeout,
			PollInterval: 5 * time.Second,
		}

		result, err := stateChangeConfig.WaitForStateContext(ctx)
		if err != nil {
			err = fmt.Errorf("error waiting for peering connection (%s) to become '%s': %v", peeringID, ps.Target, err)
			if result != nil {
				return result.(*networkmodels.HashicorpCloudNetwork20200907Peering), err
			}
			return nil, err
		}

		return result.(*networkmodels.HashicorpCloudNetwork20200907Peering), nil
	}
}

// WaitForPeeringToBePendingAcceptance will poll the GET peering endpoint until
// the state is PENDING_ACCEPTANCE, ctx is canceled, or an error occurs.
var WaitForPeeringToBePendingAcceptance = waitForPeeringToBe(peeringState{
	Target:  PeeringStatePendingAcceptance,
	Pending: []string{PeeringStateCreating},
})

// WaitForPeeringToBeAccepted will poll the GET peering endpoint until the state is ACCEPTED, ctx is canceled, or an error occurs.
var WaitForPeeringToBeAccepted = waitForPeeringToBe(peeringState{
	Target:  PeeringStateAccepted,
	Pending: []string{PeeringStateCreating, PeeringStatePendingAcceptance},
})

// WaitForPeeringToBeActive will poll the GET peering endpoint until the state is ACTIVE, ctx is canceled, or an error occurs.
var WaitForPeeringToBeActive = waitForPeeringToBe(peeringState{
	Target:  PeeringStateActive,
	Pending: WaitForPeeringToBeActiveStates,
})

// WaitForPeeringToBeActiveStates are those from which we'd expect an ACTIVE state to be possible.
var WaitForPeeringToBeActiveStates = []string{PeeringStateCreating, PeeringStatePendingAcceptance, PeeringStateAccepted}
