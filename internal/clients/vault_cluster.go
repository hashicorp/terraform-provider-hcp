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
