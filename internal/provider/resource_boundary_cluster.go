// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	boundarymodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultBoundaryClusterTimeout = time.Minute * 5

// createUpdateBoundaryClusterTimeout is the amount of time that can elapse
// before a cluster create operation should timeout.
var createBoundaryClusterTimeout = time.Minute * 25

// deleteBoundaryClusterTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteBoundaryClusterTimeout = time.Minute * 25

const boundaryClusterUpgradeTypePrefix = "UPGRADE_TYPE_"
const boundaryClusterDayOfWeekPrefix = "DAY_OF_WEEK_"

func resourceBoundaryCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to manage an HCP Boundary cluster",
		CreateContext: resourceBoundaryClusterCreate,
		UpdateContext: resourceBoundaryClusterUpdate,
		ReadContext:   resourceBoundaryClusterRead,
		DeleteContext: resourceBoundaryClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBoundaryClusterImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  &createBoundaryClusterTimeout,
			Delete:  &deleteBoundaryClusterTimeout,
			Default: &defaultBoundaryClusterTimeout,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"cluster_id": {
				Description:      "The ID of the Boundary cluster",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"username": {
				Description:      "The username of the initial admin user. This must be at least 3 characters in length, alphanumeric, hyphen, or period.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateBoundaryUsername,
			},
			"password": {
				Description:      "The password of the initial admin user. This must be at least 8 characters in length. Note that this may show up in logs, and it will be stored in the state file.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateBoundaryPassword,
				Sensitive:        true,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the Boundary cluster is located.
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Computed:     true,
			},
			// computed outputs
			"created_at": {
				Description: "The time that the Boundary cluster was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_url": {
				Description: "A unique URL identifying the Boundary cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the Boundary cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"maintenance_window_config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "The maintenance window configuration for when cluster upgrades can take place.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"upgrade_type": {
							Description:  "The upgrade type for the cluster. Valid options for upgrade type - `AUTOMATIC`, `SCHEDULED`",
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"SCHEDULED", "AUTOMATIC"}, true),
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							Default:  "AUTOMATIC",
							Optional: true,
						},
						"day": {
							Description:  "The maintenance day of the week for scheduled upgrades. Valid options for maintenance window day - `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}, true),
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							RequiredWith: []string{"maintenance_window_config.0.start"},
						},
						"start": {
							Description:  "The start time which upgrades can be performed. Uses 24H clock and must be in UTC time zone. Valid options include - 0 to 23 (inclusive)",
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 23),
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							RequiredWith: []string{"maintenance_window_config.0.day"},
						},
						"end": {
							Description:  "The end time which upgrades can be performed. Uses 24H clock and must be in UTC time zone. Valid options include - 1 to 24 (inclusive)",
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 24),
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							RequiredWith: []string{"maintenance_window_config.0.start"},
						},
					},
				},
			},
		},
	}
}

func resourceBoundaryClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	// gather the required bits to create a boundary cluster create request
	clusterID := d.Get("cluster_id").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
		// This is currently hardcoded, depending on decisions from PM
		// around regionality this may have to turn into an input
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: "aws",
			Region:   "us-east-1",
		},
	}
	upgradeType, maintenanceWindow, diagErr := getBoundaryClusterMaintainanceWindowConfig(d)
	if diagErr != nil {
		return diagErr
	}

	// check for an existing boundary cluster
	_, err = clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Boundary cluster (%s): %v", clusterID, err)
		}
		// A 404 indicates a Boundary cluster was not found.
		log.Printf("[INFO] Boundary cluster (%s) not found, proceeding with create", clusterID)
	} else {
		return diag.Errorf("a Boundary cluster with cluster_id=%q in project_id=%q already exists.", clusterID, loc.ProjectID)
	}

	// assemble the BoundaryClusterCreateRequest
	req := &boundarymodels.HashicorpCloudBoundary20211221CreateRequest{
		ClusterID: clusterID,
		Username:  username,
		Password:  password,
		Location:  loc,
	}

	// execute the Boundary cluster creation
	log.Printf("[INFO] Creating Boundary cluster (%s)", clusterID)
	createResp, err := clients.CreateBoundaryCluster(ctx, client, loc, req)
	if err != nil {
		return diag.Errorf("unable to create Boundary cluster (%s): %v", clusterID, err)
	}
	link := newLink(loc, BoundaryClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for the Boundary cluster to be created.
	if err := clients.WaitForOperation(ctx, client, "create Boundary cluster", loc, createResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create Boundary cluster (%s): %v", createResp.ClusterID, err)
	}
	log.Printf("[INFO] Created Boundary cluster (%s)", createResp.ClusterID)

	// Get the created Boundary cluster.
	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, createResp.ClusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Boundary cluster (%s): %v", createResp.ClusterID, err)
	}

	currentUpgradeType, currentMaintenanceWindow, err := clients.GetBoundaryClusterMaintenanceWindow(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve maintenance window for Boundary cluster (%s): %v", createResp.ClusterID, err)
	}

	// update the maintenance window configuration if it is passed in
	if upgradeType != nil && maintenanceWindow != nil {
		mwReq := boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindowUpdateRequest{}
		mwReq.UpgradeType = upgradeType
		mwReq.MaintenanceWindow = maintenanceWindow
		mwReq.ClusterID = cluster.ClusterID
		mwReq.Location = cluster.Location

		_, err := clients.SetBoundaryClusterMaintenanceWindow(ctx, client, loc, clusterID, &mwReq)
		if err != nil {
			return diag.Errorf("error setting maintenance window configuration for Boundary cluster (%s): %v", clusterID, err)
		}
		currentMaintenanceWindow = maintenanceWindow
		currentUpgradeType = upgradeType
	}

	// set Boundary cluster resource data
	err = setBoundaryClusterResourceData(d, cluster, currentUpgradeType, currentMaintenanceWindow)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBoundaryClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), BoundaryClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	upgradeType, maintenanceWindow, diagErr := getBoundaryClusterMaintainanceWindowConfig(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[INFO] Reading Boundary cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Boundary cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Boundary cluster (%s): %v", clusterID, err)
	}

	log.Printf("[INFO] Updating maintenance window for Boundary cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	currentUpgradeType, currentMaintenanceWindow, err := clients.GetBoundaryClusterMaintenanceWindow(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve maintenance window for Boundary cluster (%s): %v", clusterID, err)
	}

	// update the maintenance window configuration if it is created
	if upgradeType != nil && maintenanceWindow != nil {
		mwReq := boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindowUpdateRequest{}
		mwReq.UpgradeType = upgradeType
		mwReq.MaintenanceWindow = maintenanceWindow
		mwReq.ClusterID = cluster.ClusterID
		mwReq.Location = cluster.Location

		_, err := clients.SetBoundaryClusterMaintenanceWindow(ctx, client, loc, clusterID, &mwReq)
		if err != nil {
			return diag.Errorf("error setting maintenance window configuration for Boundary cluster (%s): %v", clusterID, err)
		}
		currentMaintenanceWindow = maintenanceWindow
		currentUpgradeType = upgradeType
	}

	// set Boundary cluster resource data
	err = setBoundaryClusterResourceData(d, cluster, currentUpgradeType, currentMaintenanceWindow)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBoundaryClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), BoundaryClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Boundary cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Boundary cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to fetch Boundary cluster (%s): %v", clusterID, err)
	}

	// The Boundary cluster was already deleted, remove from state.
	if *cluster.State == boundarymodels.HashicorpCloudBoundary20211221ClusterStateSTATEDELETED {
		log.Printf("[WARN] Boundary cluster (%s) failed to provision, removing from state", clusterID)
		d.SetId("")
		return nil
	}

	clusterUpgradeType, clusterMW, err := clients.GetBoundaryClusterMaintenanceWindow(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to fetch maintenenace window Boundary cluster (%s): %v", clusterID, err)
	}

	// Cluster found, update resource data.
	if err := setBoundaryClusterResourceData(d, cluster, clusterUpgradeType, clusterMW); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBoundaryClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), BoundaryClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Deleting Boundary cluster (%s)", clusterID)

	deleteResp, err := clients.DeleteBoundaryCluster(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Boundary cluster (%s) not found, so no action was taken", clusterID)
			return nil
		}

		return diag.Errorf("unable to delete Boundary cluster (%s): %v", clusterID, err)
	}

	// Wait for the delete cluster operation.
	if err := clients.WaitForOperation(ctx, client, "delete Boundary cluster", loc, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Boundary cluster (%s): %v", clusterID, err)
	}

	return nil
}

func resourceBoundaryClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_boundary_cluster.test f709ec73-55d4-46d8-897d-816ebba28778:test-boundary-cluster
	// use default project ID from provider:
	//   terraform import hcp_boundary_cluster.test test-boundary-cluster

	client := meta.(*clients.Client)
	projectID := ""
	clusterID := ""
	var err error

	if strings.Contains(d.Id(), ":") { // {project_id}:{boundary_cluster_id}
		idParts := strings.SplitN(d.Id(), ":", 2)
		clusterID = idParts[1]
		projectID = idParts[0]
	} else { // {boundary_cluster_id}
		clusterID = d.Id()
		projectID, err = GetProjectID(projectID, client.Config.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve project ID: %v", err)
		}
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: projectID,
	}

	link := newLink(loc, BoundaryClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}

func setBoundaryClusterResourceData(d *schema.ResourceData, cluster *boundarymodels.HashicorpCloudBoundary20211221Cluster, upgradeType *boundarymodels.HashicorpCloudBoundary20211221UpgradeType, clusterMW *boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindow) error {
	if err := d.Set("cluster_id", cluster.ClusterID); err != nil {
		return err
	}
	createdAtStr := cluster.CreatedAt.String()
	if err := d.Set("created_at", createdAtStr); err != nil {
		return err
	}
	if err := d.Set("cluster_url", cluster.ClusterURL); err != nil {
		return err
	}
	if err := d.Set("state", cluster.State); err != nil {
		return err
	}

	mwConfig := map[string]interface{}{}

	if upgradeType != nil {
		upgradeTypeStr := strings.TrimPrefix(string(*upgradeType), boundaryClusterUpgradeTypePrefix)
		mwConfig["upgrade_type"] = upgradeTypeStr

		if *upgradeType == boundarymodels.HashicorpCloudBoundary20211221UpgradeTypeUPGRADETYPESCHEDULED && clusterMW != nil {
			dayOfWeekStr := strings.TrimPrefix(string(*clusterMW.DayOfWeek), boundaryClusterDayOfWeekPrefix)
			mwConfig["day"] = dayOfWeekStr
			mwConfig["start"] = clusterMW.Start
			mwConfig["end"] = clusterMW.End
		} else if *upgradeType == boundarymodels.HashicorpCloudBoundary20211221UpgradeTypeUPGRADETYPESCHEDULED && clusterMW == nil {
			return fmt.Errorf("invalid maintenance window: missing configuration for SCHEDULED upgrade type")
		}
	}

	if err := d.Set("maintenance_window_config", []interface{}{mwConfig}); err != nil {
		return err
	}
	return nil
}

func getBoundaryClusterMaintainanceWindowConfig(d *schema.ResourceData) (*boundarymodels.HashicorpCloudBoundary20211221UpgradeType, *boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindow, diag.Diagnostics) {
	if !d.HasChange("maintenance_window_config") {
		return nil, nil, nil
	}

	// get the maintenance_window_config resources
	mwConfigParam, ok := d.GetOk("maintenance_window_config")
	if !ok {
		return nil, nil, nil
	}

	// convert to []interface is required even though we set a MaxItems=1
	mwConfigs, ok := mwConfigParam.([]interface{})
	if !ok || len(mwConfigs) == 0 {
		return nil, nil, nil
	}

	// get the elements in the config
	mwConfigElems, ok := mwConfigs[0].(map[string]interface{})
	if !ok || len(mwConfigElems) == 0 {
		return nil, nil, nil
	}

	upgradeTypeElem := mwConfigElems["upgrade_type"].(string)
	// add enum type prefix for type conversion
	if !strings.HasPrefix(upgradeTypeElem, boundaryClusterUpgradeTypePrefix) {
		upgradeTypeElem = boundaryClusterUpgradeTypePrefix + upgradeTypeElem
	}
	upgradeType := boundarymodels.HashicorpCloudBoundary20211221UpgradeType(upgradeTypeElem)
	maintenanceWindow := &boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindow{}

	mwDayElem := mwConfigElems["day"].(string)
	// add enum type prefix for type conversion
	if !strings.HasPrefix(mwDayElem, boundaryClusterDayOfWeekPrefix) {
		mwDayElem = boundaryClusterDayOfWeekPrefix + mwDayElem
	}
	mwDay := boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindowDayOfWeek(mwDayElem)
	mwStart := mwConfigElems["start"].(int)
	mwEnd := mwConfigElems["end"].(int)
	maintenanceWindow.DayOfWeek = &mwDay
	maintenanceWindow.Start = int32(mwStart)
	maintenanceWindow.End = int32(mwEnd)

	if upgradeType == boundarymodels.HashicorpCloudBoundary20211221UpgradeTypeUPGRADETYPESCHEDULED {
		if mwDay == "" {
			return nil, nil, diag.Errorf("maintenance window configuration is invalid: `day` is required for SCHEDULED upgrade type")
		}
		if mwStart < 0 || mwStart > 23 || mwEnd < 0 || mwEnd > 23 {
			return nil, nil, diag.Errorf("maintenance window configuration is invalid: `start` and `end` must be between 0 - 24 (inclusive) for SCHEDULED upgrade type")
		}
		if mwStart >= mwEnd {
			return nil, nil, diag.Errorf("maintenance window configuration is invalid: `start` should be less than `end` for SCHEDULED upgrade type")
		}
	} else if mwDay != "" || mwStart != 0 || mwEnd != 0 {
		return nil, nil, diag.Errorf("maintenance window configuration is invalid: `day` is only allowed on SCHEDULED upgrade type")
	}

	return &upgradeType, maintenanceWindow, nil
}
