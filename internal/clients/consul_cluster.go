package clients

import (
	"context"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	consulmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
)

// GetConsulClusterByID gets an Consul cluster by its ID
func GetConsulClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20200826Cluster, error) {

	getParams := consul_service.NewGetParams()
	getParams.Context = ctx
	getParams.ID = consulClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID
	getParams.LocationRegionProvider = &loc.Region.Provider
	getParams.LocationRegionRegion = &loc.Region.Region

	getResp, err := client.Consul.Get(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Cluster, nil
}
