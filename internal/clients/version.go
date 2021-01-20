package clients

import (
	"context"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	consulmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
)

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
