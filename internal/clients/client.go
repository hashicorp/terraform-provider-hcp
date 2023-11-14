// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cloud_billing "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"

	cloud_boundary "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client/boundary_service"

	cloud_consul "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client/consul_service"

	cloud_iam "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"

	cloud_network "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"

	cloud_operation "github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client/operation_service"

	cloud_packer "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"

	cloud_resource_manager "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"

	cloud_vault "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client/vault_service"

	cloud_vault_secrets "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"

	hcpConfig "github.com/hashicorp/hcp-sdk-go/config"
	sdk "github.com/hashicorp/hcp-sdk-go/httpclient"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Billing           billing_account_service.ClientService
	Boundary          boundary_service.ClientService
	Consul            consul_service.ClientService
	IAM               iam_service.ClientService
	Network           network_service.ClientService
	Operation         operation_service.ClientService
	Organization      organization_service.ClientService
	Packer            packer_service.ClientService
	Project           project_service.ClientService
	ServicePrincipals service_principals_service.ClientService
	Vault             vault_service.ClientService
	VaultSecrets      secret_service.ClientService
}

// ClientConfig specifies configuration for the client that interacts with HCP
type ClientConfig struct {
	ClientID       string
	ClientSecret   string
	CredentialFile string

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
	// Build the HCP Config options
	var opts []hcpConfig.HCPConfigOption
	if config.ClientID != "" && config.ClientSecret != "" {
		opts = append(opts, hcpConfig.WithClientCredentials(config.ClientID, config.ClientSecret))
	} else if config.CredentialFile != "" {
		opts = append(opts, hcpConfig.WithCredentialFilePath(config.CredentialFile))
	} else {
		opts = append(opts, hcpConfig.FromEnv())
	}

	// Create the HCP Config
	hcp, err := hcpConfig.NewHCPConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("invalid HCP config: %w", err)
	}

	// Fetch a token to verify that we have valid credentials
	if _, err := hcp.Token(); err != nil {
		return nil, fmt.Errorf("no valid credentials available: %w", err)
	}

	httpClient, err := sdk.New(sdk.Config{
		HCPConfig:     hcp,
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
		Config:            config,
		Billing:           cloud_billing.New(httpClient, nil).BillingAccountService,
		Boundary:          cloud_boundary.New(httpClient, nil).BoundaryService,
		Consul:            cloud_consul.New(httpClient, nil).ConsulService,
		IAM:               cloud_iam.New(httpClient, nil).IamService,
		Network:           cloud_network.New(httpClient, nil).NetworkService,
		Operation:         cloud_operation.New(httpClient, nil).OperationService,
		Organization:      cloud_resource_manager.New(httpClient, nil).OrganizationService,
		Packer:            cloud_packer.New(httpClient, nil).PackerService,
		Project:           cloud_resource_manager.New(httpClient, nil).ProjectService,
		ServicePrincipals: cloud_iam.New(httpClient, nil).ServicePrincipalsService,
		Vault:             cloud_vault.New(httpClient, nil).VaultService,
		VaultSecrets:      cloud_vault_secrets.New(httpClient, nil).SecretService,
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
