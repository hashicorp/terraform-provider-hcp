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
	// PeeringStateCreating is the CREATING state of a network peering
	PeeringStateCreating = string(networkmodels.HashicorpCloudNetwork20200907PeeringStateCREATING)

	// PeeringStatePendingAcceptance is the PENDING_ACCEPTANCE state of a network peering
	PeeringStatePendingAcceptance = string(networkmodels.HashicorpCloudNetwork20200907PeeringStatePENDINGACCEPTANCE)
)

// peeringRefreshState refreshes the state of the network peering by calling
// the GET endpoint
func peeringRefreshState(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		peering, err := GetPeeringByID(ctx, client, peeringID, hvnID, loc)
		if err != nil {
			return nil, "", err
		}

		return peering, string(peering.State), nil
	}
}

// WaitForPeeringToBePendingAcceptance will poll the GET peering endpoint until
// the state is PENDING_ACCEPTANCE, ctx is canceled, or an error occurs.
func WaitForPeeringToBePendingAcceptance(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907Peering, error) {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{
			PeeringStateCreating,
		},
		Target: []string{
			PeeringStatePendingAcceptance,
		},
		Refresh:      peeringRefreshState(ctx, client, peeringID, hvnID, loc),
		Timeout:      timeout,
		PollInterval: 5 * time.Second,
	}

	result, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error waiting for network peering (%s) to become 'PENDING_ACCEPTANCE': %s", peeringID, err)
	}

	return result.(*networkmodels.HashicorpCloudNetwork20200907Peering), nil
}
