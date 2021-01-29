package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetPeeringByID gets a peering by its ID, hvnID, and location
func GetPeeringByID(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907Peering, error) {
	getPeeringParams := network_service.NewGetPeeringParams()
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

// WaitForPeeringToBePendingAcceptance will poll the GET peering endpoint until
// the state is PENDING_ACCEPTANCE, ctx is canceled, or an error occurs. This
// is required because AWS can return a newly created Network peering before
// it is ready to be accepted.
// See https://hashicorp.atlassian.net/browse/HCP-1658
func WaitForPeeringToBePendingAcceptance(ctx context.Context, client *Client, peeringID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907Peering, error) {
	for {
		peering, err := GetPeeringByID(ctx, client, peeringID, hvnID, loc)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve Network peering (%s): %v", peeringID, err)
		}
		switch peering.State {
		case networkmodels.HashicorpCloudNetwork20200907PeeringStateCREATING:
			log.Printf("[INFO] Waiting for Network peering (%s) to be in PENDING_ACCEPTANCE state", peeringID)
		case networkmodels.HashicorpCloudNetwork20200907PeeringStatePENDINGACCEPTANCE:
			return peering, nil
		default:
			return nil, fmt.Errorf("unexpected peering state %s", peering.State)
		}

		// Wait some time to check the state again
		select {
		case <-time.After(time.Second * 5):
			continue
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled waiting to retrieve Network peering (%s)", peeringID)
		}
	}
}
