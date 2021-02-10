package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var tgwDefaultTimeout = time.Minute * 1
var tgwCreateTimeout = time.Minute * 35
var tgwDeleteTimeout = time.Minute * 35

func resourceAwsTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS Transit gateway attachment resource allows you to manage a Transit gateway attachment that attaches an HVN to a Transit gateway in AWS.",

		CreateContext: resourceAwsTransitGatewayAttachmentCreate,
		ReadContext:   resourceAwsTransitGatewayAttachmentRead,
		DeleteContext: resourceAwsTransitGatewayAttachmentDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &tgwDefaultTimeout,
			Create:  &tgwCreateTimeout,
			Delete:  &tgwDeleteTimeout,
		},

		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"transit_gateway_attachment_id": {
				Description:      "The user-settable name of the Transit gateway attachment in HCP.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"transit_gateway_id": {
				Description: "The ID of the Transit gateway in AWS.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"resource_share_arn": {
				Description: "The Amazon Resource Name (ARN) of the Resource Share that is needed to grant HCP access to the Transit gateway in AWS. The Resource Share should be associated with the HCP AWS account principal (see [aws_ram_principal_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_principal_association)) and the Transit gateway resource (see [aws_ram_resource_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_resource_association))",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
			},
			"destination_cidrs": {
				Description: "The list of associated CIDR ranges. Traffic from these CIDRs will be allowed for all resources in the HVN. Traffic to these CIDRs will be routed into this Transit gateway attachment.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
				Required: true,
				MinItems: 1,
				ForceNew: true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the Transit gateway attachment is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the Transit gateway attachment is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_transit_gateway_attachment_id": {
				Description: "The Transit gateway attachment ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the Transit gateway attachment was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the Transit gateway attachment will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceAwsTransitGatewayAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	tgwAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	tgwID := d.Get("transit_gateway_id").(string)
	resourceShareARN := d.Get("resource_share_arn").(string)
	rawCIDRs := d.Get("destination_cidrs").([]interface{})

	destinationCIDRs := make([]string, len(rawCIDRs))
	for i, cidr := range rawCIDRs {
		strCidr, ok := cidr.(string)
		if !ok {
			return diag.Errorf("unable to convert cidr: %v to string", cidr)
		}
		destinationCIDRs[i] = strCidr
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	// Check for an existing HVN
	_, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the Transit gateway attachment", hvnID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnID, err)
	}
	log.Printf("[INFO] HVN (%s) found, proceeding with create", hvnID)

	// Check if TGW attachment already exists
	_, err = clients.GetTGWAttachmentByID(ctx, client, tgwAttachmentID, hvnID, loc)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Transit gateway attachment (%s): %v", tgwAttachmentID, err)
		}

		log.Printf("[INFO] Transit gateway attachment (%s) not found, proceeding with create", tgwAttachmentID)
	} else {
		return diag.Errorf("a Transit gateway attachment with transit_gateway_attachment_id=%s, hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the state. Please see the resource documentation for hcp_transit_gateway_attachment for more information", tgwAttachmentID, hvnID, loc.ProjectID)
	}

	// Create TGW attachment
	createTGWAttachmentParams := network_service.NewCreateTGWAttachmentParams()
	createTGWAttachmentParams.HvnID = hvnID
	createTGWAttachmentParams.HvnLocationOrganizationID = loc.OrganizationID
	createTGWAttachmentParams.HvnLocationProjectID = loc.ProjectID
	createTGWAttachmentParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateTGWAttachmentRequest{
		Cidrs: destinationCIDRs,
		Hvn: &sharedmodels.HashicorpCloudLocationLink{
			ID:       hvnID,
			Location: loc,
		},
		ID: tgwAttachmentID,
		ProviderData: &networkmodels.HashicorpCloudNetwork20200907CreateTGWAttachmentRequestProviderData{
			AwsData: &networkmodels.HashicorpCloudNetwork20200907AWSCreateRequestTGWData{
				ResourceShareArn: resourceShareARN,
				TgwID:            tgwID,
			},
		},
	}
	log.Printf("[INFO] Creating Transit gateway attachment for HVN (%s) and Transit gateway (%s)", hvnID, tgwID)
	createTGWAttachmentResponse, err := client.Network.CreateTGWAttachment(createTGWAttachmentParams, nil)
	if err != nil {
		return diag.Errorf("unable to create Transit gateway attachment for HVN (%s) and Transit gateway (%s): %v", hvnID, tgwID, err)
	}

	tgwAtt := createTGWAttachmentResponse.Payload.TgwAttachment

	// Set the globally unique id of this TGW attachment in the state now since
	// it has been created, and from this point forward should be deletable
	link := newLink(tgwAtt.Location, TgwAttachmentResourceType, tgwAtt.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for TGW attachment creation to complete
	if err := clients.WaitForOperation(ctx, client, "create Transit gateway attachment", loc, createTGWAttachmentResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Transit gateway attachment (%s) for HVN (%s) and Transit gateway (%s): %v", tgwAtt.ID, tgwAtt.Hvn.ID, tgwAtt.ProviderData.AwsData.TgwID, err)
	}

	log.Printf("[INFO] Created Transit gateway attachment (%s) for HVN (%s) and Transit gateway (%s)", tgwAtt.ID, tgwAtt.Hvn.ID, tgwAtt.ProviderData.AwsData.TgwID)

	// Wait for TGW attachment to transition into PENDING_ACCEPTANCE state
	tgwAtt, err = clients.WaitForTGWAttachmentToBePendingAcceptance(ctx, client, tgwAtt.ID, hvnID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Transit gateway attachment (%s) is now in PENDING_ACCEPTANCE state", tgwAtt.ID)

	if err := setTransitGatewayAttachmentResourceData(d, tgwAtt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsTransitGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), TgwAttachmentResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	tgwAttID := link.ID
	loc := link.Location
	hvnID := d.Get("hvn_id").(string)

	log.Printf("[INFO] Reading Transit gateway attachment (%s)", tgwAttID)
	tgwAtt, err := clients.GetTGWAttachmentByID(ctx, client, tgwAttID, hvnID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Transit gateway attachment (%s) not found, removing from state", tgwAttID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve Transit gateway attachment (%s): %v", tgwAttID, err)
	}

	// The TGW attachment failed to provision properly so we want to let the user know and
	// remove it from state
	if tgwAtt.State == networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateFAILED {
		log.Printf("[WARN] Transit gateway attachment (%s) failed to provision, removing from state", tgwAtt.ID)
		d.SetId("")
		return nil
	}

	// TGW attachment has been found, update resource data
	if err := setTransitGatewayAttachmentResourceData(d, tgwAtt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsTransitGatewayAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), TgwAttachmentResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	tgwAttID := link.ID
	loc := link.Location
	hvnID := d.Get("hvn_id").(string)

	deleteTGWAttParams := network_service.NewDeleteTGWAttachmentParams()
	deleteTGWAttParams.ID = tgwAttID
	deleteTGWAttParams.HvnID = hvnID
	deleteTGWAttParams.HvnLocationOrganizationID = loc.OrganizationID
	deleteTGWAttParams.HvnLocationProjectID = loc.ProjectID
	log.Printf("[INFO] Deleting Transit gateway attachment (%s)", tgwAttID)
	deleteTGWAttResponse, err := client.Network.DeleteTGWAttachment(deleteTGWAttParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Transit gateway attachment (%s) not found, so no action was taken", tgwAttID)
			return nil
		}

		return diag.Errorf("unable to delete Transit gateway attachment (%s): %v", tgwAttID, err)
	}

	// Wait for TGW attachment to be deleted
	if err := clients.WaitForOperation(ctx, client, "delete Transit gateway attachment", loc, deleteTGWAttResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Transit gateway attachment (%s): %v", tgwAttID, err)
	}

	log.Printf("[INFO] Transit gateway attachment (%s) deleted, removing from state", tgwAttID)

	return nil
}

func setTransitGatewayAttachmentResourceData(d *schema.ResourceData, tgwAtt *networkmodels.HashicorpCloudNetwork20200907TGWAttachment) error {
	if err := d.Set("hvn_id", tgwAtt.Hvn.ID); err != nil {
		return err
	}
	if err := d.Set("transit_gateway_attachment_id", tgwAtt.ID); err != nil {
		return err
	}
	if err := d.Set("transit_gateway_id", tgwAtt.ProviderData.AwsData.TgwID); err != nil {
		return err
	}
	if err := d.Set("destination_cidrs", tgwAtt.Cidrs); err != nil {
		return err
	}
	if err := d.Set("organization_id", tgwAtt.Location.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", tgwAtt.Location.ProjectID); err != nil {
		return err
	}
	if err := d.Set("provider_transit_gateway_attachment_id", tgwAtt.ProviderTgwAttachmentID); err != nil {
		return err
	}
	if err := d.Set("created_at", tgwAtt.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("expires_at", tgwAtt.ExpiresAt.String()); err != nil {
		return err
	}

	return nil
}
