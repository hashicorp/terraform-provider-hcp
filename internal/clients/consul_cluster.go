package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetConsulClusterByID gets an Consul cluster by its ID
func GetConsulClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20200826Cluster, error) {

	getParams := consul_service.NewConsulServiceGetParams()
	getParams.Context = ctx
	getParams.ID = consulClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Consul.ConsulServiceGet(getParams, nil)
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

	p := consul_service.NewConsulServiceGetClientConfigParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Consul.ConsulServiceGetClientConfig(p, nil)
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

	p := consul_service.NewConsulServiceCreateCustomerMasterACLTokenParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.Body = &consulmodels.HashicorpCloudConsul20200826CreateCustomerMasterACLTokenRequest{
		ID:       consulClusterID,
		Location: loc,
	}
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Consul.ConsulServiceCreateCustomerMasterACLToken(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// CreateConsulCluster will make a call to the Consul service to initiate the create Consul
// cluster workflow.
func CreateConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulCluster *consulmodels.HashicorpCloudConsul20200826Cluster) (*consulmodels.HashicorpCloudConsul20200826CreateResponse, error) {

	p := consul_service.NewConsulServiceCreateParams()
	p.Context = ctx
	p.Body = &consulmodels.HashicorpCloudConsul20200826CreateRequest{Cluster: consulCluster}

	p.ClusterLocationOrganizationID = loc.OrganizationID
	p.ClusterLocationProjectID = loc.ProjectID

	resp, err := client.Consul.ConsulServiceCreate(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetAvailableHCPConsulVersionsForLocation gets the list of available Consul versions that HCP supports for
// the provided location.
func GetAvailableHCPConsulVersionsForLocation(ctx context.Context, loc *sharedmodels.HashicorpCloudLocationLocation, client *Client) ([]*consulmodels.HashicorpCloudConsul20200826Version, error) {
	p := consul_service.NewConsulServiceListVersionsParams()
	p.Context = ctx
	p.LocationProjectID = loc.ProjectID
	p.LocationOrganizationID = loc.OrganizationID

	resp, err := client.Consul.ConsulServiceListVersions(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// GetAvailableHCPConsulVersions gets the list of available Consul versions that HCP supports.
func GetAvailableHCPConsulVersions(ctx context.Context, client *Client) ([]*consulmodels.HashicorpCloudConsul20200826Version, error) {
	p := consul_service.NewConsulServiceListVersions2Params()
	p.Context = ctx

	resp, err := client.Consul.ConsulServiceListVersions2(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// DeleteConsulCluster will make a call to the Consul service to initiate the delete Consul
// cluster workflow.
func DeleteConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*consulmodels.HashicorpCloudConsul20200826DeleteResponse, error) {

	p := consul_service.NewConsulServiceDeleteParams()
	p.Context = ctx
	p.ID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	deleteResp, err := client.Consul.ConsulServiceDelete(p, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}

// ListConsulUpgradeVersions gets the list of available Consul versions that the supplied cluster
// can upgrade to.
func ListConsulUpgradeVersions(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) ([]*consulmodels.HashicorpCloudConsul20200826Version, error) {

	p := consul_service.NewConsulServiceListUpgradeVersionsParams()
	p.Context = ctx
	p.ID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Consul.ConsulServiceListUpgradeVersions(p, nil)

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

	updateParams := consul_service.NewConsulServiceUpdateParams()
	updateParams.Context = ctx
	updateParams.ClusterID = cluster.ID
	updateParams.ClusterLocationProjectID = loc.ProjectID
	updateParams.ClusterLocationOrganizationID = loc.OrganizationID
	updateParams.Body = &cluster

	// Invoke update cluster endpoint
	updateResp, err := client.Consul.ConsulServiceUpdate(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}
