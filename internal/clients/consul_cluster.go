// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client/consul_service"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

var (
	platformType = string(consulmodels.HashicorpCloudConsul20210204PlatformTypeHCP)
)

// GetConsulClusterByID gets an Consul cluster by its ID
func GetConsulClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20210204Cluster, error) {

	getParams := consul_service.NewGetParams()
	getParams.Context = ctx
	getParams.ID = consulClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

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
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20210204GetClientConfigResponse, error) {

	p := consul_service.NewGetClientConfigParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

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
	consulClusterID string) (*consulmodels.HashicorpCloudConsul20210204CreateCustomerMasterACLTokenResponse, error) {

	p := consul_service.NewCreateCustomerMasterACLTokenParams()
	p.Context = ctx
	p.ID = consulClusterID
	p.Body = &consulmodels.HashicorpCloudConsul20210204CreateCustomerMasterACLTokenRequest{
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
	consulCluster *consulmodels.HashicorpCloudConsul20210204Cluster) (*consulmodels.HashicorpCloudConsul20210204CreateResponse, error) {

	p := consul_service.NewCreateParams()
	p.Context = ctx
	p.Body = &consulmodels.HashicorpCloudConsul20210204CreateRequest{Cluster: consulCluster}

	p.ClusterLocationOrganizationID = loc.OrganizationID
	p.ClusterLocationProjectID = loc.ProjectID

	resp, err := client.Consul.Create(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetAvailableHCPConsulVersionsForLocation gets the list of available Consul versions that HCP supports for
// the provided location.
func GetAvailableHCPConsulVersionsForLocation(ctx context.Context, loc *sharedmodels.HashicorpCloudLocationLocation, client *Client) ([]*consulmodels.HashicorpCloudConsul20210204Version, error) {
	p := consul_service.NewListVersionsParams()
	p.Context = ctx
	p.LocationProjectID = loc.ProjectID
	p.LocationOrganizationID = loc.OrganizationID
	p.PlatformType = &platformType

	resp, err := client.Consul.ListVersions(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// GetAvailableHCPConsulVersions gets the list of available Consul versions that HCP supports.
func GetAvailableHCPConsulVersions(ctx context.Context, client *Client) ([]*consulmodels.HashicorpCloudConsul20210204Version, error) {
	p := consul_service.NewListVersions2Params()
	p.Context = ctx
	p.PlatformType = &platformType

	resp, err := client.Consul.ListVersions2(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// DeleteConsulCluster will make a call to the Consul service to initiate the delete Consul
// cluster workflow.
func DeleteConsulCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*consulmodels.HashicorpCloudConsul20210204DeleteResponse, error) {

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

// ListConsulUpgradeVersions gets the list of available Consul versions that the supplied cluster
// can upgrade to.
func ListConsulUpgradeVersions(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) ([]*consulmodels.HashicorpCloudConsul20210204Version, error) {

	p := consul_service.NewListUpgradeVersionsParams()
	p.Context = ctx
	p.ID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Consul.ListUpgradeVersions(p, nil)

	if err != nil {
		return nil, err
	}

	return resp.Payload.Versions, nil
}

// UpdateConsulCluster will make a call to the Consul service to initiate the update Consul
// cluster workflow.
func UpdateConsulCluster(ctx context.Context, client *Client,
	newCluster *consulmodels.HashicorpCloudConsul20210204Cluster) (*consulmodels.HashicorpCloudConsul20210204UpdateResponse, error) {

	updateParams := consul_service.NewUpdateParams()
	updateParams.Context = ctx
	updateParams.ClusterID = newCluster.ID
	updateParams.ClusterLocationProjectID = newCluster.Location.ProjectID
	updateParams.ClusterLocationOrganizationID = newCluster.Location.OrganizationID
	updateParams.Body = newCluster

	// Invoke update cluster endpoint
	updateResp, err := client.Consul.Update(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}
