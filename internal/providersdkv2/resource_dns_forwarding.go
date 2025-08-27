// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var dnsForwardingDefaultTimeout = time.Minute * 10

func resourceDNSForwarding() *schema.Resource {
	return &schema.Resource{
		Description:   "The DNS forwarding resource allows you to manage DNS forwarding configurations for HVNs.",
		CreateContext: resourceDNSForwardingCreate,
		ReadContext:   resourceDNSForwardingRead,
		Timeouts: &schema.ResourceTimeout{
			Create:  &dnsForwardingDefaultTimeout,
			Read:    &dnsForwardingDefaultTimeout,
			Default: &dnsForwardingDefaultTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceDNSForwardingImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HVN that this DNS forwarding belongs to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"dns_forwarding_id": {
				Description:      "The ID of the DNS forwarding configuration.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"peering_id": {
				Description: "The ID of the peering connection for DNS forwarding.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"connection_type": {
				Description: "The connection type for DNS forwarding.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"forwarding_rule": {
				Description: "The forwarding rule configuration.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id": {
							Description:      "The ID of the forwarding rule.",
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: validateSlugID,
						},
						"domain_name": {
							Description: "The domain name for DNS forwarding.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"inbound_endpoint_ips": {
							Description: "The list of inbound endpoint IP addresses.",
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the DNS forwarding is located.
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"self_link": {
				Description: "A unique URL identifying the DNS forwarding configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the DNS forwarding configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the DNS forwarding configuration was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDNSForwardingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)
	peeringID := d.Get("peering_id").(string)
	connectionType := d.Get("connection_type").(string)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	// Build HVN link
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	hvnLink := newLink(loc, HvnResourceType, hvnID)

	// Get forwarding rule configuration
	forwardingRuleList := d.Get("forwarding_rule").([]interface{})
	if len(forwardingRuleList) != 1 {
		return diag.Errorf("exactly one forwarding rule must be specified")
	}

	forwardingRuleData := forwardingRuleList[0].(map[string]interface{})
	ruleID := forwardingRuleData["rule_id"].(string)
	domainName := forwardingRuleData["domain_name"].(string)

	inboundEndpointIPsList := forwardingRuleData["inbound_endpoint_ips"].([]interface{})
	inboundEndpointIPs := make([]string, len(inboundEndpointIPsList))
	for i, ip := range inboundEndpointIPsList {
		inboundEndpointIPs[i] = ip.(string)
	}

	rule := &networkmodels.HashicorpCloudNetwork20200907ForwardingRule{
		ID:                 ruleID,
		DomainName:         domainName,
		InboundEndpointIps: inboundEndpointIPs,
	}

	log.Printf("[INFO] Creating DNS forwarding (%s) for HVN (%s)", dnsForwardingID, hvnID)
	dnsForwarding, err := client.CreateDNSForwarding(ctx, hvnID, client.Config.OrganizationID, projectID, dnsForwardingID, peeringID, connectionType, hvnLink, rule)
	if err != nil {
		return diag.Errorf("unable to create DNS forwarding (%s) for HVN (%s): %v", dnsForwardingID, hvnID, err)
	}

	link := newLink(loc, DNSForwardingResourceType, dnsForwarding.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	return resourceDNSForwardingRead(ctx, d, meta)
}

func resourceDNSForwardingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), DNSForwardingResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := link.ID

	log.Printf("[INFO] Reading DNS forwarding (%s)", dnsForwardingID)
	dnsForwarding, err := client.GetDNSForwarding(ctx, hvnID, link.Location.OrganizationID, link.Location.ProjectID, dnsForwardingID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] DNS forwarding (%s) not found, removing from state", dnsForwardingID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve DNS forwarding (%s): %v", dnsForwardingID, err)
	}

	// DNS forwarding found, update resource data.
	if err := setDNSForwardingResourceData(d, dnsForwarding, link.Location); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDNSForwardingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The import ID is expected to be in the format:
	// /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}
	// or simply {hvn_id}:{dns_forwarding_id} for brevity
	importID := d.Id()

	var projectID, hvnID, dnsForwardingID string

	if strings.HasPrefix(importID, "/project/") {
		parts := strings.Split(importID, "/")
		if len(parts) == 7 {
			projectID = parts[2]
			hvnID = parts[4]
			dnsForwardingID = parts[6]
		} else {
			return nil, fmt.Errorf("invalid import ID format, expected /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}")
		}
	} else {
		parts := strings.Split(importID, ":")
		if len(parts) == 2 {
			hvnID = parts[0]
			dnsForwardingID = parts[1]
		} else {
			return nil, fmt.Errorf("invalid import ID format, expected hvn_id:dns_forwarding_id or /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}")
		}
	}

	client := meta.(*clients.Client)
	if projectID == "" {
		projectID = client.Config.ProjectID
	}

	d.Set("project_id", projectID)
	d.Set("hvn_id", hvnID)

	// Build the resource ID using the DNS forwarding ID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	link := newLink(loc, DNSForwardingResourceType, dnsForwardingID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}
	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}

// setDNSForwardingResourceData sets the KV pairs of the DNS forwarding resource schema.
func setDNSForwardingResourceData(d *schema.ResourceData, dnsForwarding *networkmodels.HashicorpCloudNetwork20200907DNSForwardingResponse, loc *sharedmodels.HashicorpCloudLocationLocation) error {
	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}

	if err := d.Set("dns_forwarding_id", dnsForwarding.ID); err != nil {
		return err
	}

	if err := d.Set("peering_id", dnsForwarding.PeeringID); err != nil {
		return err
	}

	if err := d.Set("connection_type", dnsForwarding.ConnectionType); err != nil {
		return err
	}

	// Set forwarding rule
	if len(dnsForwarding.Rules) > 0 && dnsForwarding.Rules[0].Rule != nil {
		rule := dnsForwarding.Rules[0]
		forwardingRule := map[string]interface{}{
			"rule_id":              rule.Rule.ID,
			"domain_name":          rule.Rule.DomainName,
			"inbound_endpoint_ips": rule.Rule.InboundEndpointIps,
		}
		if err := d.Set("forwarding_rule", []interface{}{forwardingRule}); err != nil {
			return err
		}
	}

	link := newLink(loc, DNSForwardingResourceType, dnsForwarding.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	if dnsForwarding.State != nil {
		if err := d.Set("state", string(*dnsForwarding.State)); err != nil {
			return err
		}
	}

	if err := d.Set("created_at", dnsForwarding.CreatedAt.String()); err != nil {
		return err
	}

	return nil
}
