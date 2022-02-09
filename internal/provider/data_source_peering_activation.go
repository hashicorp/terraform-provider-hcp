package provider

import (
	"context"
	"log"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var activationDefaultTimeout = time.Minute * 5

// The purpose of this data source is to ensure the given peering connection is active before the Terraform run returns. Since
// peering connections generally must return early with some outputs in order to complete the peering, it may take some time before the peering connection is active.
// This data source polls the peering connection and returns once it verifies the state is active.
func dataSourcePeeringActivation() *schema.Resource {
	return &schema.Resource{
		Description: "The peering connection activation data source allows you to verify the active state of an existing peering connection.",

		ReadContext: dataSourcePeeringActivationRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &activationDefaultTimeout,
		},

		Schema: map[string]*schema.Schema{
			// Required inputs
			"peering_id": {
				Description: "The ID of the peering connection.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hvn_link": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func dataSourcePeeringActivationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	peeringID := d.Get("peering_id").(string)
	orgID := client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	_, err = clients.GetPeeringByID(ctx, client, peeringID, hvnLink.ID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Peering connection (%s) not found, removing activation data source from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	peering, err := clients.WaitForPeeringToBeActive(ctx, client, peeringID, hvnLink.ID, loc, activationDefaultTimeout)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Peering connection (%s) is now in ACTIVE state", peering.ID)

	d.SetId(d.Get("peering_id").(string))

	return nil
}
