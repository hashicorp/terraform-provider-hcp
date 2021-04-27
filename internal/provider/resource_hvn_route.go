package provider

import (
	"time"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var hvnRouteDefaultTimeout = time.Minute * 1

func setHVNRouteResourceData(d *schema.ResourceData, route *networkmodels.HashicorpCloudNetwork20200907HVNRoute) error {
	if err := d.Set("hvn_id", route.Hvn.ID); err != nil {
		return err
	}

	if err := d.Set("organization_id", route.Hvn.Location.OrganizationID); err != nil {
		return err
	}

	if err := d.Set("project_id", route.Hvn.Location.ProjectID); err != nil {
		return err
	}

	if err := d.Set("hvn_route_id", route.ID); err != nil {
		return err
	}

	if err := d.Set("destination_cidr", route.Destination); err != nil {
		return err
	}

	if err := d.Set("target", route.Target); err != nil {
		return err
	}

	if err := d.Set("state", route.State); err != nil {
		return err
	}

	if err := d.Set("created_at", route.CreatedAt.String()); err != nil {
		return err
	}

	return nil
}
