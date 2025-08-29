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
func (c Client) GetDNSForwardingRule(ctx context.Context, hvnID, organizationID, projectID, dnsForwardingID, ruleID string) (*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {
	params := network_service.NewGetDNSForwardingRuleParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID
	params.ID = ruleID

	resp, err := c.Network.GetDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwardingRule, nil
}

// CreateDNSForwardingRule creates a DNS forwarding rule.
func (c Client) CreateDNSForwardingRule(ctx context.Context, hvnID, organizationID, projectID, dnsForwardingID string, rule *networkmodels.HashicorpCloudNetwork20200907ForwardingRule, hvnLink *sharedmodels.HashicorpCloudLocationLink) (*networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleResponse, error) {
	params := network_service.NewCreateDNSForwardingRuleParams()
	params.Context = ctx
	params.DNSForwardingRuleHvnID = hvnID
	params.DNSForwardingRuleHvnLocationOrganizationID = organizationID
	params.DNSForwardingRuleHvnLocationProjectID = projectID
	params.DNSForwardingRuleDNSForwardingID = dnsForwardingID

	params.Body = &networkmodels.HashicorpCloudNetwork20200907CreateDNSForwardingRuleRequest{
		DNSForwardingRule: &networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule{
			Hvn:  hvnLink,
			Rule: rule,
		},
	}

	resp, err := c.Network.CreateDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DeleteDNSForwardingRule deletes a DNS forwarding rule.
func (c Client) DeleteDNSForwardingRule(ctx context.Context, hvnID, organizationID, projectID, dnsForwardingID, ruleID string) (*networkmodels.HashicorpCloudNetwork20200907DeleteDNSForwardingRuleResponse, error) {
	params := network_service.NewDeleteDNSForwardingRuleParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.LocationOrganizationID = organizationID
	params.LocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID
	params.ID = ruleID

	resp, err := c.Network.DeleteDNSForwardingRule(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// ListDNSForwardingRules lists DNS forwarding rules for a DNS forwarding.
func (c Client) ListDNSForwardingRules(ctx context.Context, hvnID, organizationID, projectID, dnsForwardingID string) ([]*networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, error) {
	params := network_service.NewListDNSForwardingRulesParams()
	params.Context = ctx
	params.HvnID = hvnID
	params.HvnLocationOrganizationID = organizationID
	params.HvnLocationProjectID = projectID
	params.DNSForwardingID = dnsForwardingID

	resp, err := c.Network.ListDNSForwardingRules(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.DNSForwardingRules, nil
}
