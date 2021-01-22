package clients

import (
	"context"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	consulmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
)

// CreateSnapshot  will make a call to the Consul service to initiate the create Consul
// snapshot workflow.
func CreateSnapshot(ctx context.Context, client *Client, res *sharedmodels.HashicorpCloudLocationLink,
	snapshotName string) (*consulmodels.HashicorpCloudConsul20200826CreateSnapshotResponse, error) {

	p := consul_service.NewCreateSnapshotParams()
	p.Context = ctx
	p.ResourceLocationOrganizationID = res.Location.OrganizationID
	p.ResourceLocationProjectID = res.Location.ProjectID
	p.Body = &consulmodels.HashicorpCloudConsul20200826CreateSnapshotRequest{
		Name:     snapshotName,
		Resource: res,
	}

	resp, err := client.Consul.CreateSnapshot(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetSnapshotByID gets an Consul snapshot by its ID
func GetSnapshotByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	snapshotID string) (*consulmodels.HashicorpCloudConsul20200826GetSnapshotResponse, error) {

	p := consul_service.NewGetSnapshotParams()
	p.Context = ctx
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID
	p.SnapshotID = snapshotID

	resp, err := client.Consul.GetSnapshot(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
