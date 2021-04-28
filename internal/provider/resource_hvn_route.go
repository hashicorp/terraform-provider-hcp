package provider

import (
	"errors"
	"time"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var hvnRouteDefaultTimeout = time.Minute * 1

func setHVNRouteResourceData(d *schema.ResourceData, route *networkmodels.HashicorpCloudNetwork20200907HVNRoute,
	loc *sharedmodels.HashicorpCloudLocationLocation) error {

	// Set self_link for the HVN route.
	link := newLink(loc, HVNRouteResourceType, route.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}

	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	// Set self_link identifying the target of the HVN route.
	var targetLink string

	switch route.Target.HvnConnection.Type {
	case HvnResourceType:
		hvnLink := newLink(loc, HvnResourceType, route.Target.HvnConnection.ID)
		targetLink, err = linkURL(hvnLink)
		if err != nil {
			return err
		}
	case PeeringResourceType:
		peeringLink := newLink(loc, PeeringResourceType, route.Target.HvnConnection.ID)
		targetLink, err = linkURL(peeringLink)
		if err != nil {
			return err
		}
	case TgwAttachmentResourceType:
		tgwAttLink := newLink(loc, PeeringResourceType, route.Target.HvnConnection.ID)
		targetLink, err = linkURL(tgwAttLink)
		if err != nil {
			return err
		}
	default:
		return errors.New("Unable to set self_link identifying the target - HVN Route target is not a known type.")
	}

	if err := d.Set("target", map[string]interface{}{"self_link": targetLink}); err != nil {
		return err
	}

	if err := d.Set("destination_cidr", route.Destination); err != nil {
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
