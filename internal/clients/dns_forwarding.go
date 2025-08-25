// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// CreateDNSForwarding creates a new DNS Forwarding within an HVN
func CreateDNSForwarding(ctx context.Context, client *Client,
	id string,
	hvn *sharedmodels.HashicorpCloudLocationLink,
	dnsForwardingID string,
	peeringID string,
	ruleID string,
	domainName string,
	inboundEndpointIps []string,
	connectionType string,
	location *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingResponse, error) {

	dnsForwardingParams := network_service.NewCreateDNSForwardingParams()
	dnsForwardingParams.Context = ctx
	dnsForwardingParams.DNSForwardingHvnLocationOrganizationID = location.OrganizationID
	dnsForwardingParams.DNSForwardingHvnLocationProjectID = location.ProjectID
	dnsForwardingParams.DNSForwardingHvnID = hvn.ID
	dnsForwardingParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRequest{
		DNSForwarding: &networkmodels.HashicorpCloudNetwork20200907DNSForwarding{
			ConnectionType: connectionType,
			ID:             dnsForwardingID,
			Hvn:            hvn,
			PeeringID:      peeringID,
			Rule: &networkmodels.HashicorpCloudNetwork20200907ForwardingRule{
				ID:                 ruleID,
				DomainName:         domainName,
				InboundEndpointIps: inboundEndpointIps,
			},
		},
	}

	log.Printf("[INFO] Creating DNS Forwarding with ID %s and rule with ID %s for HVN (%s) with peering %s", dnsForwardingID, ruleID, hvn.ID, peeringID)
	dnsForwardingResp, err := client.Network.CreateDNSForwarding(dnsForwardingParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create DNS Forwarding with ID %s and rule with ID %s for HVN (%s) with peering %s: %v", hvn.ID, peeringID, err)
	}

	return dnsForwardingResp.Payload, nil
}

// GetDNSForwarding returns specific DNS Forwarding by its ID
func GetDNSForwarding(ctx context.Context, client *Client, hvnID, dnsForwardingID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907DNSForwardingResponse, error) {
	getDNSForwardingParams := network_service.NewGetDNSForwardingParams()
	getDNSForwardingParams.Context = ctx
	getDNSForwardingParams.HvnID = hvnID
	getDNSForwardingParams.ID = dnsForwardingID
	getDNSForwardingParams.HvnLocationOrganizationID = loc.OrganizationID
	getDNSForwardingParams.HvnLocationProjectID = loc.ProjectID

	getDNSForwardingResponse, err := client.Network.GetDNSForwarding(getDNSForwardingParams, nil)
	if err != nil {
		return nil, err
	}

	return getDNSForwardingResponse.Payload.DNSForwarding, nil
}

// ListDNSForwardings lists the dns forwardings for an HVN.
func ListDNSForwardings(ctx context.Context, client *Client, hvnID string,
	loc *sharedmodels.HashicorpCloudLocationLocation) ([]*networkmodels.HashicorpCloudNetwork20200907DNSForwardingResponse, error) {

	listDNSForwardingsParams := network_service.NewListDNSForwardingsParams()
	listDNSForwardingsParams.Context = ctx
	listDNSForwardingsParams.HvnID = hvnID
	listDNSForwardingsParams.HvnLocationOrganizationID = loc.OrganizationID
	listDNSForwardingsParams.HvnLocationProjectID = loc.ProjectID

	listDNSForwardingsResponse, err := client.Network.ListDNSForwardings(listDNSForwardingsParams, nil)
	if err != nil {
		return nil, err
	}

	return listDNSForwardingsResponse.Payload.DNSForwardings, nil
}
