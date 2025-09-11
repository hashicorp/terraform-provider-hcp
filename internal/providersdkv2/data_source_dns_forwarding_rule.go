package providersdkv2

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceDNSForwardingRule() *schema.Resource {
	return &schema.Resource{
		Description: "The DNS forwarding rule data source provides information about an existing DNS forwarding rule.",
		ReadContext: dataSourceDNSForwardingRuleRead,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HVN that this DNS forwarding rule belongs to.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"dns_forwarding_id": {
				Description:      "The ID of the DNS forwarding configuration this rule belongs to.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"dns_forwarding_rule_id": {
				Description:      "The ID of the DNS forwarding rule.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the DNS forwarding rule is located.
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"domain_name": {
				Description: "The domain name for which DNS forwarding rule was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"inbound_endpoint_ips": {
				Description: "The IP addresses of the target customer network inbound endpoints to which the DNS requests for the above domain will be forwarded.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"self_link": {
				Description: "A unique URL identifying the DNS forwarding rule.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the DNS forwarding rule.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the DNS forwarding rule was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceDNSForwardingRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)
	dnsForwardingRuleID := d.Get("dns_forwarding_rule_id").(string)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	log.Printf("[INFO] Reading DNS forwarding rule (%s) in DNS forwarding (%s)", dnsForwardingRuleID, dnsForwardingID)
	rule, err := clients.GetDNSForwardingRule(ctx, client, hvnID, client.Config.OrganizationID, projectID, dnsForwardingID, dnsForwardingRuleID)
	if err != nil {
		return diag.Errorf("unable to retrieve DNS forwarding rule (%s): %v", dnsForwardingRuleID, err)
	}

	// Set computed fields
	if err := d.Set("project_id", projectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain_name", rule.Rule.DomainName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("inbound_endpoint_ips", rule.Rule.InboundEndpointIps); err != nil {
		return diag.FromErr(err)
	}
	if rule.State != nil {
		if err := d.Set("state", string(*rule.State)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("created_at", rule.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	// Create self link
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	link := newLink(loc, DNSForwardingRuleResourceType, rule.Rule.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("self_link", url); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	return nil
}
