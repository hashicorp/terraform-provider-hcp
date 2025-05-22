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

// CreateDNSForwardingRule creates a new DNS Forwarding Rule for an existing DNS Forwarding in an HVN.
func CreateDNSForwardingRule(ctx context.Context, client *Client,
	id string,
	hvn *sharedmodels.HashicorpCloudLocationLink,
	dnsForwardingID string,
	// peeringID string,
	dnsForwardingRuleID string,
	domainName string,
	inboundEndpointIps []string,
	location *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleResponse, error) {

	dnsForwardingRuleParams := network_service.NewCreateDNSForwardingRuleParams()
	dnsForwardingRuleParams.Context = ctx
	dnsForwardingRuleParams.DNSForwardingRuleHvnLocationOrganizationID = location.OrganizationID
	dnsForwardingRuleParams.DNSForwardingRuleHvnLocationOrganizationID = location.ProjectID
	dnsForwardingRuleParams.DNSForwardingRuleHvnID = hvn.ID
	dnsForwardingRuleParams.DNSForwardingRuleDNSForwardingID = dnsForwardingID
	dnsForwardingRuleParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleRequest{
		DNSForwardingRule: &networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule{
			DNSForwardingID: dnsForwardingID,
			Hvn:             hvn,
			Rule: &networkmodels.HashicorpCloudNetwork20200907ForwardingRule{
				ID:                 dnsForwardingRuleID,
				DomainName:         domainName,
				InboundEndpointIps: inboundEndpointIps,
			},
		},
	}

	log.Printf("[INFO] Creating DNS Forwarding Rule with ID %s in DNS Forwarding ID %s for HVN (%s)", dnsForwardingRuleID, dnsForwardingID, hvn.ID)
	dnsForwardingRuleResp, err := client.Network.CreateDNSForwardingRule(dnsForwardingRuleParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create DNS Forwarding Rule with ID %s in DNS Forwarding ID %s for HVN (%s): %v", dnsForwardingRuleID, dnsForwardingID, hvn.ID, err)
	}

	return dnsForwardingRuleResp.Payload, nil
}

// GetDNSForwardingRule returns specific DNS Forwarding Rule by its ID in a DNS Forwarding.
func GetDNSForwardingRule(ctx context.Context, client *Client, hvnID, dnsForwardingID, dnsForwardinRuleID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {
	getDNSForwardingRuleParams := network_service.NewGetDNSForwardingRuleParams()
	getDNSForwardingRuleParams.Context = ctx
	getDNSForwardingRuleParams.HvnID = hvnID
	getDNSForwardingRuleParams.ID = dnsForwardinRuleID
	getDNSForwardingRuleParams.DNSForwardingID = dnsForwardingID
	getDNSForwardingRuleParams.HvnLocationOrganizationID = loc.OrganizationID
	getDNSForwardingRuleParams.HvnLocationProjectID = loc.ProjectID

	getDNSForwardingRuleResponse, err := client.Network.GetDNSForwardingRule(getDNSForwardingRuleParams, nil)
	if err != nil {
		return nil, err
	}

	return getDNSForwardingRuleResponse.Payload.DNSForwardingRule, nil
}

// ListDNSForwardingRules lists the dns forwarding rules for a specific dns forwarding.
func ListDNSForwardingRules(ctx context.Context, client *Client, hvnID, dnsForwardingID string,
	loc *sharedmodels.HashicorpCloudLocationLocation) ([]*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {

	listDNSForwardingRulesParams := network_service.NewListDNSForwardingRulesParams()
	listDNSForwardingRulesParams.Context = ctx
	listDNSForwardingRulesParams.HvnID = hvnID
	listDNSForwardingRulesParams.HvnLocationOrganizationID = loc.OrganizationID
	listDNSForwardingRulesParams.HvnLocationProjectID = loc.ProjectID
	listDNSForwardingRulesParams.DNSForwardingID = dnsForwardingID

	listDNSForwardingRulesResponse, err := client.Network.ListDNSForwardingRules(listDNSForwardingRulesParams, nil)
	if err != nil {
		return nil, err
	}

	return listDNSForwardingRulesResponse.Payload.DNSForwardingRules, nil
}

// DeleteDNSForwardingRule deletes a DNS Forwarding Rule by its ID in a DNS Forwarding.
func DeleteDNSForwardingRule(ctx context.Context, client *Client, hvnID, dnsForwardingID, dnsForwardingRuleID string, loc *sharedmodels.HashicorpCloudLocationLocation) (*networkmodels.HashicorpCloudNetwork20200907DeleteDNSForwardingRuleResponse, error) {
	deleteDNSForwardingRuleParams := network_service.NewDeleteDNSForwardingRuleParams()
	deleteDNSForwardingRuleParams.Context = ctx
	deleteDNSForwardingRuleParams.HvnID = hvnID
	deleteDNSForwardingRuleParams.ID = dnsForwardingRuleID
	deleteDNSForwardingRuleParams.DNSForwardingID = dnsForwardingID
	deleteDNSForwardingRuleParams.LocationOrganizationID = loc.OrganizationID
	deleteDNSForwardingRuleParams.LocationProjectID = loc.ProjectID

	deleteDNSForwardingRuleResp, err := client.Network.DeleteDNSForwardingRule(deleteDNSForwardingRuleParams, nil)
	if err != nil {
		return nil, err
	}

	return deleteDNSForwardingRuleResp.Payload, nil
}
