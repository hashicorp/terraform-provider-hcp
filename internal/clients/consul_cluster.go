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

// GetConsulClientConfigFiles gets a Consul cluster set of client config files.
//
// The files will be returned in base64-encoded format and will get passed in
// that format.
func GetConsulClientConfigFiles(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20200826GetClientConfigResponse, error) {

	p := consul_service.NewGetClientConfigParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID
	p.LocationRegionProvider = &loc.Region.Provider
	p.LocationRegionRegion = &loc.Region.Region

	resp, err := client.Consul.GetClientConfig(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// CreateCustomerRootACLToken invokes the consul-service endpoint to create
// privileged tokens for a Consul cluster.
// Example token: After cluster create, a customer would want a root token
// (or "bootstrap token") so they can continue to set-up their cluster.
func CreateCustomerRootACLToken(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20200826CreateCustomerMasterACLTokenResponse, error) {

	p := consul_service.NewCreateCustomerMasterACLTokenParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.Body = &consulmodels.HashicorpCloudConsul20200826CreateCustomerMasterACLTokenRequest{
		ID:       consulClusterID,
		Location: loc,
	}
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Consul.CreateCustomerMasterACLToken(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// CreateConsulCluster will make a call to the Consul service to initiate the create Consul
// cluster workflow.
func CreateConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID, datacenter, consulVersion string, numServers int32, private, connectEnabled bool, network *sharedmodels.HashicorpCloudLocationLink) (*consulmodels.HashicorpCloudConsul20200826CreateResponse, error) {

	p := consul_service.NewCreateParams()
	p.Context = ctx
	p.Body = &consulmodels.HashicorpCloudConsul20200826CreateRequest{
		Cluster: &consulmodels.HashicorpCloudConsul20200826Cluster{
			Config: &consulmodels.HashicorpCloudConsul20200826ClusterConfig{
				CapacityConfig: &consulmodels.HashicorpCloudConsul20200826CapacityConfig{
					NumServers: numServers,
				},
				ConsulConfig: &consulmodels.HashicorpCloudConsul20200826ConsulConfig{
					ConnectEnabled: connectEnabled,
					Datacenter:     datacenter,
				},
				MaintenanceConfig: nil,
				NetworkConfig: &consulmodels.HashicorpCloudConsul20200826NetworkConfig{
					Network: network,
					Private: private,
				},
			},
			ConsulVersion: consulVersion,
			ID:            clusterID,
			Location:      loc,
		},
	}

	p.ClusterLocationOrganizationID = loc.OrganizationID
	p.ClusterLocationProjectID = loc.ProjectID

	resp, err := client.Consul.Create(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetAvailableHCPConsulVersions gets the list of available Consul versions that HCP supports.
func GetAvailableHCPConsulVersions(ctx context.Context, client *Client) ([]*consulmodels.HashicorpCloudConsul20200826Version, error) {
	p := consul_service.NewListVersionsParams()
	p.Context = ctx

	resp, err := client.Consul.ListVersions(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// DeleteConsulCluster will make a call to the Consul service to initiate the delete Consul
// cluster workflow.
func DeleteConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*consulmodels.HashicorpCloudConsul20200826DeleteResponse, error) {

	p := consul_service.NewDeleteParams()
	p.Context = ctx
	p.ID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	deleteResp, err := client.Consul.Delete(p, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}

// GetAvailableHCPConsulVersions gets the list of available Consul versions that HCP supports.
func ListConsulUpgradeVersions(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) ([]*consulmodels.HashicorpCloudConsul20200826Version, error) {

	p := consul_service.NewListUpgradeVersionsParams()
	p.Context = ctx
	p.ID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID
	p.LocationRegionProvider = &loc.Region.Provider
	p.LocationRegionRegion = &loc.Region.Region

	resp, err := client.Consul.ListUpgradeVersions(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// UpdateConsulCluster will make a call to the Consul service to initiate the update Consul
// cluster workflow.
func UpdateConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID, newConsulVersion string) (*consulmodels.HashicorpCloudConsul20200826UpdateResponse, error) {

	cluster := consulmodels.HashicorpCloudConsul20200826Cluster{
		ConsulVersion: newConsulVersion,
		ID:            clusterID,
		Location: &sharedmodels.HashicorpCloudLocationLocation{
			ProjectID:      loc.ProjectID,
			OrganizationID: loc.OrganizationID,
			Region: &sharedmodels.HashicorpCloudLocationRegion{
				Region:   loc.Region.Region,
				Provider: loc.Region.Provider,
			},
		},
	}

	updateParams := consul_service.NewUpdateParams()
	updateParams.Context = ctx
	updateParams.ClusterID = cluster.ID
	updateParams.ClusterLocationProjectID = loc.ProjectID
	updateParams.ClusterLocationOrganizationID = loc.OrganizationID
	updateParams.Body = &cluster

	// Invoke update cluster endpoint
	updateResp, err := client.Consul.Update(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}
