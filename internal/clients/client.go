// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/auth/workload"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cloud_billing "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"

	cloud_boundary "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client/boundary_service"

	cloud_consul "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/client/consul_service"

	cloud_iam "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"

	cloud_network "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"

	cloud_operation "github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client/operation_service"

	cloud_packer "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"

	cloud_packer_v2 "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client"
	packer_service_v2 "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"

	cloud_resource_manager "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	cloud_vault "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client/vault_service"

	cloud_vault_secrets "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"

	cloud_waypoint "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"

	cloud_log_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/log_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/streaming_service"

	cloud_webhook "github.com/hashicorp/hcp-sdk-go/clients/cloud-webhook/stable/2023-05-31/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-webhook/stable/2023-05-31/client/webhook_service"

	cloud_vault_radar "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client"
	radar_src_registration_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/data_source_registration_service"
	radar_connection_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_connection_service"
	radar_subscription_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_subscription_service"
	radar_resource_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/resource_service"

	hcpConfig "github.com/hashicorp/hcp-sdk-go/config"
	hcpConfigGeography "github.com/hashicorp/hcp-sdk-go/config/geography"
	sdk "github.com/hashicorp/hcp-sdk-go/httpclient"
)

const (
	// TODO: replace these two with the constants as defined in
	// hcp-sdk-go/config/geography
	GeographyUS = "us"
	GeographyEU = "eu"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Billing                        billing_account_service.ClientService
	Boundary                       boundary_service.ClientService
	Consul                         consul_service.ClientService
	IAM                            iam_service.ClientService
	Network                        network_service.ClientService
	Operation                      operation_service.ClientService
	Organization                   organization_service.ClientService
	Packer                         packer_service.ClientService
	PackerV2                       packer_service_v2.ClientService
	Project                        project_service.ClientService
	ServicePrincipals              service_principals_service.ClientService
	Groups                         groups_service.ClientService
	Vault                          vault_service.ClientService
	VaultSecrets                   secret_service.ClientService
	Waypoint                       waypoint_service.ClientService
	Webhook                        webhook_service.ClientService
	LogService                     log_service.ClientService
	LogStreamingService            streaming_service.ClientService
	ResourceService                resource_service.ClientService
	RadarSourceRegistrationService radar_src_registration_service.ClientService
	RadarConnectionService         radar_connection_service.ClientService
	RadarSubscriptionService       radar_subscription_service.ClientService
	RadarResourceService           radar_resource_service.ClientService
}

// ClientConfig specifies configuration for the client that interacts with HCP
type ClientConfig struct {
	ClientID       string
	ClientSecret   string
	CredentialFile string
	Geography      string

	// WorkloadIdentityTokenFile and WorkloadIdentityResourceName can be set to
	// indicate that authentication should occur by using workload identity
	// federation. WorkloadIdentityTokenFile indicates the token and
	// WorkloadIdentityResourceName is the workload identity provider resource
	// name to authenticate against.
	//
	// Alternatively, WorkloadIdentityToken can be set to the token directly. It
	// is an error to set both WorkloadIdentityTokenFile and
	// WorkloadIdentityToken.
	WorkloadIdentityTokenFile    string
	WorkloadIdentityToken        string
	WorkloadIdentityResourceName string

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
	opts := []hcpConfig.HCPConfigOption{hcpConfig.FromEnv()}
	if config.ClientID != "" && config.ClientSecret != "" {
		opts = append(opts, hcpConfig.WithClientCredentials(config.ClientID, config.ClientSecret))
	} else if config.CredentialFile != "" {
		opts = append(opts, hcpConfig.WithCredentialFilePath(config.CredentialFile))
	} else if cf := loadCredentialFile(config); cf != nil {
		opts = append(opts, hcpConfig.WithCredentialFile(cf))
	}

	// TODO: once the naming in hcp-sdk-go has been finalized, the following can
	// be changed to
	// 		if config.Geography != "" {
	//			opts = append(opts, hcpConfig.WithGeography(config.Geography))
	// 		}
	if config.Geography == GeographyUS {
		opts = append(opts, hcpConfig.WithGeography(string(hcpConfigGeography.US)))
	} else if config.Geography == GeographyEU {
		opts = append(opts, hcpConfig.WithGeography(string(hcpConfigGeography.EU)))
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
		Config:                         config,
		Billing:                        cloud_billing.New(httpClient, nil).BillingAccountService,
		Boundary:                       cloud_boundary.New(httpClient, nil).BoundaryService,
		Consul:                         cloud_consul.New(httpClient, nil).ConsulService,
		IAM:                            cloud_iam.New(httpClient, nil).IamService,
		Network:                        cloud_network.New(httpClient, nil).NetworkService,
		Operation:                      cloud_operation.New(httpClient, nil).OperationService,
		Organization:                   cloud_resource_manager.New(httpClient, nil).OrganizationService,
		Packer:                         cloud_packer.New(httpClient, nil).PackerService,
		PackerV2:                       cloud_packer_v2.New(httpClient, nil).PackerService,
		Project:                        cloud_resource_manager.New(httpClient, nil).ProjectService,
		ServicePrincipals:              cloud_iam.New(httpClient, nil).ServicePrincipalsService,
		Groups:                         cloud_iam.New(httpClient, nil).GroupsService,
		Vault:                          cloud_vault.New(httpClient, nil).VaultService,
		VaultSecrets:                   cloud_vault_secrets.New(httpClient, nil).SecretService,
		Waypoint:                       cloud_waypoint.New(httpClient, nil).WaypointService,
		LogService:                     cloud_log_service.New(httpClient, nil).LogService,
		LogStreamingService:            cloud_log_service.New(httpClient, nil).StreamingService,
		Webhook:                        cloud_webhook.New(httpClient, nil).WebhookService,
		ResourceService:                cloud_resource_manager.New(httpClient, nil).ResourceService,
		RadarSourceRegistrationService: cloud_vault_radar.New(httpClient, nil).DataSourceRegistrationService,
		RadarConnectionService:         cloud_vault_radar.New(httpClient, nil).IntegrationConnectionService,
		RadarSubscriptionService:       cloud_vault_radar.New(httpClient, nil).IntegrationSubscriptionService,
		RadarResourceService:           cloud_vault_radar.New(httpClient, nil).ResourceService,
	}

	return client, nil
}

// loadCredentialFile loads the credential file from the given config. If the
// config does not specify workload identity authentication, this function
// returns nil.
func loadCredentialFile(config ClientConfig) *auth.CredentialFile {
	// hasWorkloadIdentityToken is true if the config specified a direct token.
	hasWorkloadIdentityToken := config.WorkloadIdentityToken != ""
	// hasWorkloadIdentityTokenFile is true if the config specified a path to a
	// file that contains the token.
	hasWorkloadIdentityTokenFile := config.WorkloadIdentityTokenFile != ""
	// hasWorkloadIdentityResource is true if the config specified a resource
	// name to authenticate against.
	hasWorkloadIdentityResource := config.WorkloadIdentityResourceName != ""

	// Overall, we consider workload identity authentication to be enabled if we
	// have a token from either source (direct or within a file) and a resource
	// name.
	hasWorkloadIdentity := (hasWorkloadIdentityToken || hasWorkloadIdentityTokenFile) && hasWorkloadIdentityResource

	if !hasWorkloadIdentity {
		return nil
	}

	switch {
	case hasWorkloadIdentityToken:
		// The direct token takes priority over the file path in case both
		// are set, so we'll check that first.
		return &auth.CredentialFile{
			Scheme: auth.CredentialFileSchemeWorkload,
			Workload: &workload.IdentityProviderConfig{
				ProviderResourceName: config.WorkloadIdentityResourceName,
				Token: &workload.CredentialTokenSource{
					Token: config.WorkloadIdentityToken,
				},
			},
		}
	default:
		// If we don't have the token directly, fall back to checking the
		// file. We checked earlier that at least one of the two options
		// were set, so if the token wasn't set the file information must
		// be present.
		return &auth.CredentialFile{
			Scheme: auth.CredentialFileSchemeWorkload,
			Workload: &workload.IdentityProviderConfig{
				ProviderResourceName: config.WorkloadIdentityResourceName,
				File: &workload.FileCredentialSource{
					Path: config.WorkloadIdentityTokenFile,
				},
			},
		}
	}
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

func (cl *Client) GetOrganizationID() string {
	if cl == nil {
		return ""
	}
	return cl.Config.OrganizationID
}

func (cl *Client) GetProjectID() string {
	if cl == nil {
		return ""
	}
	return cl.Config.ProjectID
}

// Location returns the organization and project ID to use for a given resource
// The resource project ID takes precedence over the provider project ID
func (cl *Client) Location(resourceProjectID types.String) (string, string) {
	orgID := cl.Config.OrganizationID
	projID := cl.Config.ProjectID

	// project ID defined in the resource schema has precedence over the project ID from the provider
	if !resourceProjectID.IsUnknown() {
		projID = resourceProjectID.ValueString()
	}

	return orgID, projID
}
