// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetDNSForwarding gets a DNS forwarding by its ID.
func GetDNSForwarding(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID string) (*networkmodels.HashicorpCloudNetwork20200907DNSForwardingResponse, error) {
	params := network_service.NewGetDNSForwardingParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID
	params.ID = dnsForwardingID

	resp, err := client.Network.GetDNSForwarding(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwarding, nil
}

// CreateDNSForwarding creates a DNS forwarding.
func CreateDNSForwarding(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID, peeringID, connectionType string, hvnLink *sharedmodels.HashicorpCloudLocationLink, rule *networkmodels.HashicorpCloudNetwork20200907ForwardingRule) (*networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingResponse, error) {
	params := network_service.NewCreateDNSForwardingParams()
	params.Context = ctx
	params.DNSForwardingHvnID = hvnID
	params.DNSForwardingHvnLocationOrganizationID = organizationID
	params.DNSForwardingHvnLocationProjectID = projectID

	dnsForwarding := &networkmodels.HashicorpCloudNetwork20200907DNSForwarding{
		ID:             dnsForwardingID,
		Hvn:            hvnLink,
		PeeringID:      peeringID,
		ConnectionType: connectionType,
		Rule:           rule,
	}

	params.Body = &networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRequest{
		DNSForwarding: dnsForwarding,
	}

	resp, err := client.Network.CreateDNSForwarding(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// ListDNSForwardings lists DNS forwardings for a HVN.
func ListDNSForwardings(ctx context.Context, client *Client, hvnID, organizationID, projectID string) ([]*networkmodels.HashicorpCloudNetwork20200907DNSForwardingResponse, error) {
	params := network_service.NewListDNSForwardingsParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID

	resp, err := client.Network.ListDNSForwardings(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwardings, nil
}
