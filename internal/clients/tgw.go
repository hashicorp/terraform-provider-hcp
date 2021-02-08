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

// GetTGWAttachmentByID gets a TGW attachment by its ID, hvnID, and location
func GetTGWAttachmentByID(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907TGWAttachment, error) {
	getTGWAttachmentParams := network_service.NewGetTGWAttachmentParams()
	getTGWAttachmentParams.ID = tgwAttachmentID
	getTGWAttachmentParams.HvnID = hvnID
	getTGWAttachmentParams.HvnLocationOrganizationID = loc.OrganizationID
	getTGWAttachmentParams.HvnLocationProjectID = loc.ProjectID
	getTGWAttachmentResponse, err := client.Network.GetTGWAttachment(getTGWAttachmentParams, nil)
	if err != nil {
		return nil, err
	}

	return getTGWAttachmentResponse.Payload.TgwAttachment, nil
}

// WaitForTGWAttachmentState will poll the GET TGW attachment endpoint until
// it is in the specified state, ctx is canceled, or an error occurs. This
// is required because AWS can return a newly created TGW attachment before
// it is ready to be used.
func WaitForTGWAttachmentState(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, state string) (*networkmodels.HashicorpCloudNetwork20200907TGWAttachment, error) {
	for {
		tgwAtt, err := GetTGWAttachmentByID(ctx, client, tgwAttachmentID, hvnID, loc)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve Transit gateway attachment (%s): %v", tgwAttachmentID, err)
		}
		switch curState := string(tgwAtt.State); curState {
		case state:
			return tgwAtt, nil
		case string(networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateFAILED):
			return nil, fmt.Errorf("Transit gateway attachment got into FAILED state")
		default:
			log.Printf("[INFO] Waiting for Transit gateway attachment (%s) to be in [%s] state; current state: [%s]", tgwAttachmentID, state, curState)
		}

		// Wait some time to check the state again
		select {
		case <-time.After(time.Second * 5):
			continue
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled waiting to retrieve Transit gateway attachment (%s)", tgwAttachmentID)
		}
	}
}
