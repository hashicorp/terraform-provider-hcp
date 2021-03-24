package clients

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/preview/2020-11-25/client/vault_service"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/preview/2020-11-25/models"
)

// GetVaultClusterByID gets an Vault cluster by its ID
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
