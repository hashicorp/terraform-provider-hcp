// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetDNSForwardingRule gets a DNS forwarding rule by its ID.
func GetDNSForwardingRule(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID, ruleID string) (*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {
	params := network_service.NewGetDNSForwardingRuleParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID
	params.ID = ruleID

	resp, err := client.Network.GetDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwardingRule, nil
}

// CreateDNSForwardingRule creates a DNS forwarding rule.
func CreateDNSForwardingRule(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID string, rule *networkmodels.HashicorpCloudNetwork20200907ForwardingRule, hvnLink *sharedmodels.HashicorpCloudLocationLink) (*networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleResponse, error) {
	params := network_service.NewCreateDNSForwardingRuleParams()
	params.Context = ctx
	params.DNSForwardingRuleHvnID = hvnID
	params.DNSForwardingRuleHvnLocationOrganizationID = organizationID
	params.DNSForwardingRuleHvnLocationProjectID = projectID
	params.DNSForwardingRuleDNSForwardingID = dnsForwardingID

	dnsForwardingRule := &networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule{
		Hvn:  hvnLink,
		Rule: rule,
	}

	params.Body = &networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleRequest{
		DNSForwardingRule: dnsForwardingRule,
	}

	resp, err := client.Network.CreateDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DeleteDNSForwardingRule deletes a DNS forwarding rule by its ID.
func DeleteDNSForwardingRule(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID, ruleID string) (*networkmodels.HashicorpCloudNetwork20200907DeleteDNSForwardingRuleResponse, error) {
	params := network_service.NewDeleteDNSForwardingRuleParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.LocationOrganizationID = organizationID
	params.LocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID
	params.ID = ruleID

	resp, err := client.Network.DeleteDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// ListDNSForwardingRules lists DNS forwarding rules for a DNS forwarding configuration.
func ListDNSForwardingRules(ctx context.Context, client *Client, hvnID, organizationID, projectID, dnsForwardingID string) ([]*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {
	params := network_service.NewListDNSForwardingRulesParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID

	resp, err := client.Network.ListDNSForwardingRules(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwardingRules, nil
}
