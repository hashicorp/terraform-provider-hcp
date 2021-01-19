package clients

import (
	"net/http"

	sdk "github.com/hashicorp/cloud-sdk-go"
	cloud_consul "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	cloud_network "github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/client"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	cloud_operation "github.com/hashicorp/cloud-sdk-go/clients/cloud-operation/preview/2020-05-05/client"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-operation/preview/2020-05-05/client/operation_service"
	cloud_resource_manager "github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Network      network_service.ClientService
	Operation    operation_service.ClientService
	Project      project_service.ClientService
	Organization organization_service.ClientService
	Consul       consul_service.ClientService
}

// ClientConfig specifies configuration for the client that interacts with HCP
type ClientConfig struct {
	ClientID     string
	ClientSecret string

	// OrganizationID (optional) is the organization unique identifier to launch resources in.
	OrganizationID string

	// ProjectID (optional) is the project unique identifier to launch resources in.
	ProjectID string

	// SourceChannel denotes the client (channel) that originated the HCP cluster request.
	// this is synonymous to a user-agent.
	SourceChannel string
}

// NewClient creates a new Client that is capable of making HCP requests
func NewClient(config ClientConfig) (*Client, error) {
	httpClient, err := sdk.New(sdk.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	})
	if err != nil {
		return nil, err
	}

	rt := WithHeader(httpClient.Transport)
	rt.Set("X-HCP-SOURCE-CHANNEL", config.SourceChannel)
	httpClient.Transport = rt

	client := &Client{
		Config: config,

		Network:      cloud_network.New(httpClient, nil).NetworkService,
		Operation:    cloud_operation.New(httpClient, nil).OperationService,
		Project:      cloud_resource_manager.New(httpClient, nil).ProjectService,
		Organization: cloud_resource_manager.New(httpClient, nil).OrganizationService,
		Consul:       cloud_consul.New(httpClient, nil).ConsulService,
	}

	return client, nil
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}
	return h.rt.RoundTrip(req)
}
