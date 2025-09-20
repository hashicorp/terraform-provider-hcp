// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceDNSForwarding() *schema.Resource {
	return &schema.Resource{
		Description: "The DNS forwarding data source provides information about a DNS forwarding configuration for a HashiCorp Virtual Network (HVN).",
		ReadContext: dataSourceDNSForwardingRead,
		Timeouts: &schema.ResourceTimeout{
			Read: &dnsForwardingDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description: "The ID of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"dns_forwarding_id": {
				Description: "The ID of the DNS forwarding.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional inputs
			"project_id": {
				Description: "The ID of the HCP project where the DNS forwarding is located. " +
					"If not specified, the project specified in the HCP Provider config block will be used, if configured. " +
					"If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the DNS forwarding is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peering_id": {
				Description: "The ID of the peering connection or transit gateway attachment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connection_type": {
				Description: "The type of connection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"forwarding_rules": {
				Description: "The forwarding rules for the DNS forwarding.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id": {
							Description: "The ID of the forwarding rule.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"domain_name": {
							Description: "The domain name for DNS forwarding.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"inbound_endpoint_ips": {
							Description: "The list of inbound endpoint IP addresses.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"self_link": {
				Description: "A unique URL identifying the DNS forwarding configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the DNS forwarding.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the DNS forwarding was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The time that the DNS forwarding was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceDNSForwardingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}
	organizationID := client.Config.OrganizationID

	dnsForwarding, err := clients.GetDNSForwarding(ctx, client, hvnID, organizationID, projectID, dnsForwardingID)
	if err != nil {
		return diag.Errorf("unable to fetch DNS forwarding (%s): %v", dnsForwardingID, err)
	}

	// Set the DNS forwarding data.
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	if err := d.Set("organization_id", organizationID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("peering_id", dnsForwarding.PeeringID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("connection_type", dnsForwarding.ConnectionType); err != nil {
		return diag.FromErr(err)
	}

	// Set forwarding rules for data source (plural)
	if len(dnsForwarding.Rules) > 0 {
		rules := make([]interface{}, len(dnsForwarding.Rules))
		for i, rule := range dnsForwarding.Rules {
			if rule.Rule != nil {
				rules[i] = map[string]interface{}{
					"rule_id":              rule.Rule.ID,
					"domain_name":          rule.Rule.DomainName,
					"inbound_endpoint_ips": rule.Rule.InboundEndpointIps,
				}
			}
		}
		if err := d.Set("forwarding_rules", rules); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set other computed fields
	link := newLink(loc, DNSForwardingResourceType, dnsForwarding.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return diag.FromErr(err)
	}

	if dnsForwarding.State != nil {
		if err := d.Set("state", string(*dnsForwarding.State)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("created_at", dnsForwarding.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dnsForwarding.ID)

	return nil
}
