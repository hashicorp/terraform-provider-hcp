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

var dnsForwardingRuleDefaultTimeout = time.Minute * 10

func resourceDNSForwardingRule() *schema.Resource {
	return &schema.Resource{
		Description:   "The DNS forwarding rule resource allows you to manage DNS forwarding rules for HVNs.",
		CreateContext: resourceDNSForwardingRuleCreate,
		ReadContext:   resourceDNSForwardingRuleRead,
		DeleteContext: resourceDNSForwardingRuleDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  &dnsForwardingRuleDefaultTimeout,
			Read:    &dnsForwardingRuleDefaultTimeout,
			Delete:  &dnsForwardingRuleDefaultTimeout,
			Default: &dnsForwardingRuleDefaultTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceDNSForwardingRuleImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HVN that this DNS forwarding rule belongs to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"dns_forwarding_id": {
				Description:      "The ID of the DNS forwarding configuration this rule belongs to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"domain_name": {
				Description: "The domain name for which DNS forwarding rule needs to be created.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"inbound_endpoint_ips": {
				Description: "The IP addresses of the target customer network inbound endpoints to which the DNS requests for the above domain will be forwarded.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the DNS forwarding rule is located.
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

func resourceDNSForwardingRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)
	domainName := d.Get("domain_name").(string)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	// Build location
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	// Get the HVN to obtain region information
	hvn, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		return diag.Errorf("unable to find existing HVN (%s): %v", hvnID, err)
	}

	// Update location with region information from HVN
	loc.Region = &sharedmodels.HashicorpCloudLocationRegion{
		Provider: hvn.Location.Region.Provider,
		Region:   hvn.Location.Region.Region,
	}

	inboundEndpointIPs := make([]string, 0)
	for _, ip := range d.Get("inbound_endpoint_ips").([]interface{}) {
		inboundEndpointIPs = append(inboundEndpointIPs, ip.(string))
	}

	// Get the HVN to obtain region information
	hvn, hvnErr := clients.GetHvnByID(ctx, client, loc, hvnID)
	if hvnErr != nil {
		return diag.Errorf("unable to find existing HVN (%s): %v", hvnID, hvnErr)
	}

	// Update location with region information from HVN
	loc.Region = &sharedmodels.HashicorpCloudLocationRegion{
		Provider: hvn.Location.Region.Provider,
		Region:   hvn.Location.Region.Region,
	}

	// Build HVN link with complete location
	hvnLink := newLink(loc, HvnResourceType, hvnID)

	rule := &networkmodels.HashicorpCloudNetwork20200907ForwardingRule{
		DomainName:         domainName,
		InboundEndpointIps: inboundEndpointIPs,
	}

	log.Printf("[INFO] Creating DNS forwarding rule (%s) in DNS forwarding (%s)", domainName, dnsForwardingID)
	createResp, err := client.CreateDNSForwardingRule(ctx, hvnID, client.Config.OrganizationID, projectID, dnsForwardingID, rule, hvnLink)
	if err != nil {
		return diag.Errorf("unable to create DNS forwarding rule (%s): %v", domainName, err)
	}

	link := newLink(loc, DNSForwardingRuleResourceType, createResp.DNSForwardingRule.Rule.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// Wait for the DNS forwarding rule to be created
	if err := clients.WaitForOperation(ctx, client, "create DNS forwarding rule", loc, createResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create DNS forwarding rule (%s): %v", createResp.DNSForwardingRule.Rule.ID, err)
	}

	log.Printf("[INFO] Created DNS forwarding rule (%s)", createResp.DNSForwardingRule.Rule.ID)

	// Get the updated DNS forwarding rule
	dnsRule, ruleErr := client.GetDNSForwardingRule(ctx, hvnID, client.Config.OrganizationID, projectID, dnsForwardingID, createResp.DNSForwardingRule.Rule.ID)
	if ruleErr != nil {
		return diag.Errorf("unable to retrieve DNS forwarding rule (%s): %v", createResp.DNSForwardingRule.Rule.ID, ruleErr)
	}

	if err := setDNSForwardingRuleResourceData(d, dnsRule, loc); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDNSForwardingRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), DNSForwardingRuleResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)
	ruleID := link.ID

	log.Printf("[INFO] Reading DNS forwarding rule (%s) in DNS forwarding (%s)", ruleID, dnsForwardingID)
	rule, err := client.GetDNSForwardingRule(ctx, hvnID, link.Location.OrganizationID, link.Location.ProjectID, dnsForwardingID, ruleID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] DNS forwarding rule (%s) not found, removing from state", ruleID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve DNS forwarding rule (%s): %v", ruleID, err)
	}

	// DNS forwarding rule found, update resource data.
	if err := setDNSForwardingRuleResourceData(d, rule, link.Location); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDNSForwardingRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), DNSForwardingRuleResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvnID := d.Get("hvn_id").(string)
	dnsForwardingID := d.Get("dns_forwarding_id").(string)
	ruleID := link.ID

	log.Printf("[INFO] Deleting DNS forwarding rule (%s) in DNS forwarding (%s)", ruleID, dnsForwardingID)
	deleteResp, err := client.DeleteDNSForwardingRule(ctx, hvnID, link.Location.OrganizationID, link.Location.ProjectID, dnsForwardingID, ruleID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] DNS forwarding rule (%s) not found during delete, removing from state", ruleID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to delete DNS forwarding rule (%s): %v", ruleID, err)
	}

	// Wait for the DNS forwarding rule to be deleted
	if err := clients.WaitForOperation(ctx, client, "delete DNS forwarding rule", link.Location, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete DNS forwarding rule (%s): %v", ruleID, err)
	}

	log.Printf("[INFO] DNS forwarding rule (%s) deleted successfully", ruleID)

	return nil
}

func resourceDNSForwardingRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The import ID is expected to be in the format:
	// /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}/dns-forwarding-rule/{rule_id}
	// or simply {hvn_id}:{dns_forwarding_id}:{rule_id} for brevity
	importID := d.Id()

	var projectID, hvnID, dnsForwardingID, ruleID string

	if strings.HasPrefix(importID, "/project/") {
		parts := strings.Split(importID, "/")
		if len(parts) == 9 {
			projectID = parts[2]
			hvnID = parts[4]
			dnsForwardingID = parts[6]
			ruleID = parts[8]
		} else {
			return nil, fmt.Errorf("invalid import ID format, expected /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}/dns-forwarding-rule/{rule_id}")
		}
	} else {
		parts := strings.Split(importID, ":")
		if len(parts) == 3 {
			hvnID = parts[0]
			dnsForwardingID = parts[1]
			ruleID = parts[2]
		} else {
			return nil, fmt.Errorf("invalid import ID format, expected hvn_id:dns_forwarding_id:rule_id or /project/{project_id}/hvn/{hvn_id}/dns-forwarding/{dns_forwarding_id}/dns-forwarding-rule/{rule_id}")
		}
	}

	client := meta.(*clients.Client)
	if projectID == "" {
		projectID = client.Config.ProjectID
	}

	d.Set("project_id", projectID)
	d.Set("hvn_id", hvnID)
	d.Set("dns_forwarding_id", dnsForwardingID)

	// Build the resource ID using the rule ID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	link := newLink(loc, DNSForwardingRuleResourceType, ruleID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}
	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}

// setDNSForwardingRuleResourceData sets the KV pairs of the DNS forwarding rule resource schema.
func setDNSForwardingRuleResourceData(d *schema.ResourceData, rule *networkmodels.HashicorpCloudNetwork20200907DNSForwardingRule, loc *sharedmodels.HashicorpCloudLocationLocation) error {
	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}

	link := newLink(loc, DNSForwardingRuleResourceType, rule.Rule.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	if err := d.Set("domain_name", rule.Rule.DomainName); err != nil {
		return err
	}

	if err := d.Set("inbound_endpoint_ips", rule.Rule.InboundEndpointIps); err != nil {
		return err
	}

	if rule.State != nil {
		if err := d.Set("state", string(*rule.State)); err != nil {
			return err
		}
	}

	if err := d.Set("created_at", rule.CreatedAt.String()); err != nil {
		return err
	}

	return nil
}
