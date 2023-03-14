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

// GetTGWAttachmentByID gets a TGW attachment by its ID, hvnID, and location
func GetTGWAttachmentByID(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907TGWAttachment, error) {
	getTGWAttachmentParams := network_service.NewGetTGWAttachmentParams()
	getTGWAttachmentParams.Context = ctx
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

const (
	// TgwAttachmentStateCreating is the CREATING state of a TGW attachment
	TgwAttachmentStateCreating = string(networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateCREATING)

	// TgwAttachmentStatePendingAcceptance is the PENDING_ACCEPTANCE state of a TGW attachment
	TgwAttachmentStatePendingAcceptance = string(networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStatePENDINGACCEPTANCE)

	// TgwAttachmentStateAccepted is the ACCEPTED state of a TGW attachment
	TgwAttachmentStateAccepted = string(networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateACCEPTED)

	// TgwAttachmentStateActive is the ACTIVE state of a TGW attachment
	TgwAttachmentStateActive = string(networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateACTIVE)
)

// tgwAttachmentRefreshState refreshes the state of the TGW attachment by
// calling the GET endpoint
func tgwAttachmentRefreshState(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		tgwAtt, err := GetTGWAttachmentByID(ctx, client, tgwAttachmentID, hvnID, loc)
		if err != nil {
			return nil, "", err
		}

		return tgwAtt, string(tgwAtt.State), nil
	}
}

// WaitForTGWAttachmentToBeActive will poll the GET TGW attachment endpoint
// until the state is ACTIVE, ctx is canceled, or an error occurs.
func WaitForTGWAttachmentToBeActive(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907TGWAttachment, error) {
	stateChangeConf := resource.StateChangeConf{
		Pending: WaitForTGWAttachmentToBeActiveStates,
		Target: []string{
			TgwAttachmentStateActive,
		},
		Refresh:      tgwAttachmentRefreshState(ctx, client, tgwAttachmentID, hvnID, loc),
		Timeout:      timeout,
		PollInterval: 5 * time.Second,
	}

	result, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		err = fmt.Errorf("error waiting for transit gateway attachment (%s) to become 'ACTIVE': %s", tgwAttachmentID, err)
		if result != nil {
			return result.(*networkmodels.HashicorpCloudNetwork20200907TGWAttachment), err
		}
		return nil, err
	}

	return result.(*networkmodels.HashicorpCloudNetwork20200907TGWAttachment), nil
}

// WaitForTGWAttachmentToBeActiveStates is the set of states of the attachment which we'll wait on.
var WaitForTGWAttachmentToBeActiveStates = []string{
	TgwAttachmentStateCreating,
	TgwAttachmentStatePendingAcceptance,
	TgwAttachmentStateAccepted,
}

// WaitForTGWAttachmentToBePendingAcceptance will poll the GET TGW attachment
// endpoint until the state is PENDING_ACCEPTANCE, ctx is canceled, or an error
// occurs.
func WaitForTGWAttachmentToBePendingAcceptance(ctx context.Context, client *Client, tgwAttachmentID string, hvnID string, loc *sharedmodels.HashicorpCloudLocationLocation, timeout time.Duration) (*networkmodels.HashicorpCloudNetwork20200907TGWAttachment, error) {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{
			TgwAttachmentStateCreating,
		},
		Target: []string{
			TgwAttachmentStatePendingAcceptance,
		},
		Refresh:      tgwAttachmentRefreshState(ctx, client, tgwAttachmentID, hvnID, loc),
		Timeout:      timeout,
		PollInterval: 5 * time.Second,
	}

	result, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for transit gateway attachment (%s) to become 'PENDING_ACCEPTANCE': %s", tgwAttachmentID, err)
	}

	return result.(*networkmodels.HashicorpCloudNetwork20200907TGWAttachment), nil
}
