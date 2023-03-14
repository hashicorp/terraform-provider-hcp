// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"errors"
	"fmt"
	"log"
	"strings"

	cloud_network "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	cloud_operation "github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client/operation_service"
	cloud_resource_manager "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cloud_consul "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client/consul_service"

	cloud_vault "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client/vault_service"

	cloud_packer "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"

	cloud_boundary "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client/boundary_service"

	sdk "github.com/hashicorp/hcp-sdk-go/httpclient"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Network      network_service.ClientService
	Operation    operation_service.ClientService
	Project      project_service.ClientService
	Organization organization_service.ClientService
	Consul       consul_service.ClientService
	Vault        vault_service.ClientService
	Packer       packer_service.ClientService
	Boundary     boundary_service.ClientService
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
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		SourceChannel: config.SourceChannel,
	})
	if err != nil {
		return nil, err
	}

	httpClient.SetLogger(logger{})

	if ShouldLog() {
		httpClient.Debug = true
	}

	client := &Client{
		Config: config,

		Network:      cloud_network.New(httpClient, nil).NetworkService,
		Operation:    cloud_operation.New(httpClient, nil).OperationService,
		Project:      cloud_resource_manager.New(httpClient, nil).ProjectService,
		Organization: cloud_resource_manager.New(httpClient, nil).OrganizationService,
		Consul:       cloud_consul.New(httpClient, nil).ConsulService,
		Vault:        cloud_vault.New(httpClient, nil).VaultService,
		Packer:       cloud_packer.New(httpClient, nil).PackerService,
		Boundary:     cloud_boundary.New(httpClient, nil).BoundaryService,
	}

	return client, nil
}

type providerMeta struct {
	ModuleName string `cty:"module_name"`
}

// updateSourceChannel updates the SourceChannel of the client
func (cl *Client) UpdateSourceChannel(d *schema.ResourceData) (*Client, error) {

	// Adds module metadata if any.
	var m providerMeta

	err := d.GetProviderMeta(&m)
	if err != nil {
		return cl, errors.New("failed to get provider meta")
	}

	if m.ModuleName != "" {
		sc := cl.Config.SourceChannel
		sc = strings.Join([]string{sc, fmt.Sprintf("terraform-module/%s", m.ModuleName)}, " ")
		cl.Config.SourceChannel = sc

		// Return a new client with the updated source channel
		cl, err = NewClient(cl.Config)
		if err != nil {
			log.Printf("failed to create new client with updated source channel: %v", err)
		}
	}

	return cl, nil
}
