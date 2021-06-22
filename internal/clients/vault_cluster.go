package clients

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/preview/2020-11-25/client/vault_service"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/preview/2020-11-25/models"
)

// GetVaultClusterByID gets an Vault cluster by its ID.
func GetVaultClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultClusterID string) (*vaultmodels.HashicorpCloudVault20201125Cluster, error) {

	getParams := vault_service.NewGetParams()
	getParams.Context = ctx
	getParams.ClusterID = vaultClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Vault.Get(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Cluster, nil
}

// CreateVaultCluster will make a call to the Consul service to initiate the create Consul
// cluster workflow.
func CreateVaultCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultCluster *vaultmodels.HashicorpCloudVault20201125InputCluster) (*vaultmodels.HashicorpCloudVault20201125CreateResponse, error) {

	p := vault_service.NewCreateParams()
	p.Context = ctx
	p.Body = &vaultmodels.HashicorpCloudVault20201125CreateRequest{Cluster: vaultCluster}

	p.ClusterLocationOrganizationID = loc.OrganizationID
	p.ClusterLocationProjectID = loc.ProjectID

	resp, err := client.Vault.Create(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DeleteVaultCluster will make a call to the Vault service to initiate the delete Vault
// cluster workflow.
func DeleteVaultCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*vaultmodels.HashicorpCloudVault20201125DeleteResponse, error) {

	p := vault_service.NewDeleteParams()
	p.Context = ctx
	p.ClusterID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	deleteResp, err := client.Vault.Delete(p, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}

// CreateVaultClusterAdminToken will make a call to the Vault service to generate an admin token for the Vault cluster
// that expires after 6 hours.
func CreateVaultClusterAdminToken(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultClusterID string) (*vaultmodels.HashicorpCloudVault20201125GetAdminTokenResponse, error) {

	p := vault_service.NewGetAdminTokenParams()
	p.Context = ctx
	p.ClusterID = vaultClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID
	p.LocationRegionProvider = &loc.Region.Provider
	p.LocationRegionRegion = &loc.Region.Region

	resp, err := client.Vault.GetAdminToken(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// UpdateVaultClusterPublicIps will make a call to the Vault service to enable or disable public IPs for the Vault cluster.
func UpdateVaultClusterPublicIps(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string, enablePublicIps bool) (*vaultmodels.HashicorpCloudVault20201125UpdatePublicIpsResponse, error) {

	updateParams := vault_service.NewUpdatePublicIpsParams()
	updateParams.Context = ctx
	updateParams.ClusterID = clusterID
	updateParams.LocationProjectID = loc.ProjectID
	updateParams.LocationOrganizationID = loc.OrganizationID
	updateParams.Body = &vaultmodels.HashicorpCloudVault20201125UpdatePublicIpsRequest{
		// ClusterID and Location are repeated because the values above are required to populate the URL,
		// and the values below are required in the API request body
		ClusterID:       clusterID,
		Location:        loc,
		EnablePublicIps: enablePublicIps,
	}

	updateResp, err := client.Vault.UpdatePublicIps(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}
