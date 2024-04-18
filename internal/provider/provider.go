// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/iam"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/logstreaming"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/resourcemanager"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/vaultsecrets"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/webhook"
)

// This is an implementation using the Provider framework
// Docs can be found here: https://developer.hashicorp.com/terraform/plugin/framework
// NOTE: All other resources and data sources for other products can be found in the
// providersdkv2 folder at the same level
type ProviderFramework struct {
	version string
}

type ProviderFrameworkConfiguration struct {
	Client *clients.Client
}

type ProviderFrameworkModel struct {
	ClientSecret     types.String `tfsdk:"client_secret"`
	ClientID         types.String `tfsdk:"client_id"`
	CredentialFile   types.String `tfsdk:"credential_file"`
	ProjectID        types.String `tfsdk:"project_id"`
	WorkloadIdentity types.List   `tfsdk:"workload_identity"`
}

type WorkloadIdentityFrameworkModel struct {
	TokenFile    types.String `tfsdk:"token_file"`
	ResourceName types.String `tfsdk:"resource_name"`
}

func (p *ProviderFramework) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hcp"
	resp.Version = "dev"
}

func (p *ProviderFramework) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth2 Client ID for API operations.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("client_secret")),
					stringvalidator.ConflictsWith(path.MatchRoot("credential_file")),
					stringvalidator.ConflictsWith(path.MatchRoot("workload_identity")),
				},
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth2 Client Secret for API operations.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("client_id")),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The default project in which resources should be created.",
			},
			"credential_file": schema.StringAttribute{
				Optional: true,
				Description: "The path to an HCP credential file to use to authenticate the provider to HCP. " +
					"You can alternatively set the HCP_CRED_FILE environment variable to point at a credential file as well. " +
					"Using a credential file allows you to authenticate the provider as a service principal via client " +
					"credentials or dynamically based on Workload Identity Federation.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("workload_identity")),
				},
			},
		},
		Blocks: map[string]schema.Block{
			// TODO migrate to SingleNestedAttribute once the providersdkv2 is
			// fully migrated.
			"workload_identity": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"token_file": schema.StringAttribute{
							Required:    true,
							Description: "The path to a file containing a JWT token retrieved from an OpenID Connect (OIDC) or OAuth2 provider.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"resource_name": schema.StringAttribute{
							Required:    true,
							Description: "The resource_name of the Workload Identity Provider to exchange the token with.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^iam/project/.+/service-principal/.+/workload-identity-provider/.+$`),
									"must be a workload identity provider resource_name",
								),
							},
						},
					},
				},
				Description: "Allows authenticating the provider by exchanging the OAuth 2.0 access token or OpenID Connect " +
					"token specified in the `token_file` for a HCP service principal using Workload Identity Federation.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
			},
		},
	}
}

func (p *ProviderFramework) Resources(ctx context.Context) []func() resource.Resource {
	return append([]func() resource.Resource{
		// Resource Manager
		resourcemanager.NewOrganizationIAMPolicyResource,
		resourcemanager.NewOrganizationIAMBindingResource,

		resourcemanager.NewProjectResource,
		resourcemanager.NewProjectIAMPolicyResource,
		resourcemanager.NewProjectIAMBindingResource,
		// Vault Secrets
		vaultsecrets.NewVaultSecretsAppResource,
		vaultsecrets.NewVaultSecretsSecretResource,
		// IAM
		iam.NewServicePrincipalResource,
		iam.NewServicePrincipalKeyResource,
		iam.NewWorkloadIdentityProviderResource,
		iam.NewGroupResource,
		iam.NewGroupMembersResource,
		// Log Streaming
		logstreaming.NewHCPLogStreamingDestinationResource,
		// Webhook
		webhook.NewNotificationsWebhookResource,
		// Waypoint
		waypoint.NewActionConfigResource,
		waypoint.NewApplicationResource,
		waypoint.NewApplicationTemplateResource,
		waypoint.NewAddOnResource,
		waypoint.NewAddOnDefinitionResource,
		waypoint.NewTfcConfigResource,
	}, packer.ResourceSchemaBuilders...)
}

func (p *ProviderFramework) DataSources(ctx context.Context) []func() datasource.DataSource {
	return append([]func() datasource.DataSource{
		// Resource Manager
		resourcemanager.NewProjectDataSource,
		resourcemanager.NewOrganizationDataSource,
		resourcemanager.NewIAMPolicyDataSource,
		// Vault Secrets
		vaultsecrets.NewVaultSecretsAppDataSource,
		vaultsecrets.NewVaultSecretsSecretDataSource,
		// IAM
		iam.NewServicePrincipalDataSource,
		iam.NewGroupDataSource,
		iam.NewUserPrincipalDataSource,
		// Waypoint
		waypoint.NewActionConfigDataSource,
		waypoint.NewApplicationDataSource,
		waypoint.NewApplicationTemplateDataSource,
		waypoint.NewAddOnDataSource,
		waypoint.NewAddOnDefinitionDataSource,
	}, packer.DataSourceSchemaBuilders...)
}

func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProviderFramework{
			version: version,
		}
	}
}

func (p *ProviderFramework) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// In order to avoid disrupting testing and development, the HCP status check only runs on prod.
	// HCP_API_HOST is used to point the provider at test environments. When unset, the provider points to prod.
	if os.Getenv("HCP_API_HOST") == "" || os.Getenv("HCP_API_HOST") == "api.cloud.hashicorp.com" {
		// This helper verifies HCP's status and either returns a warning for degraded performance
		// or errors out if there's an outage.
		resp.Diagnostics.Append(isHCPOperationalFramework()...)
	}

	// Sets up HCP SDK client.
	var data ProviderFrameworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	clientConfig := clients.ClientConfig{
		ClientID:       data.ClientID.ValueString(),
		ClientSecret:   data.ClientSecret.ValueString(),
		CredentialFile: data.CredentialFile.ValueString(),
		ProjectID:      data.ProjectID.ValueString(),
		SourceChannel:  "terraform-provider-hcp",
	}

	// Read the workload_identity configuration.
	if len(data.WorkloadIdentity.Elements()) == 1 {
		elements := make([]WorkloadIdentityFrameworkModel, 0, 1)
		resp.Diagnostics.Append(data.WorkloadIdentity.ElementsAs(ctx, &elements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		clientConfig.WorloadIdentityTokenFile = elements[0].TokenFile.ValueString()
		clientConfig.WorkloadIdentityResourceName = elements[0].ResourceName.ValueString()
	}

	client, err := clients.NewClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("unable to create HCP api client: %v", err), "")
		return
	}

	// Attempt to source from the environment if unset.
	if clientConfig.ProjectID == "" {
		clientConfig.ProjectID = os.Getenv("HCP_PROJECT_ID")
	}

	if clientConfig.ProjectID != "" {
		getProjParams := project_service.NewProjectServiceGetParams()
		getProjParams.ID = clientConfig.ProjectID
		project, err := clients.RetryProjectServiceGet(client, getProjParams)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("unable to fetch project %q: %v", clientConfig.ProjectID, err), "")
			return
		}

		client.Config.ProjectID = project.Payload.Project.ID
		client.Config.OrganizationID = project.Payload.Project.Parent.ID

	} else {
		// For the initial release of the HCP TFP, since only one project was allowed per organization at the time,
		// the provider handled used the single organization's single project by default, instead of requiring the
		// user to set it. Once multiple projects are available, this helper issues a warning: when multiple projects exist within the org,
		// a project ID should be set on the provider or on each resource. Otherwise, the oldest project will be used by default.
		// This helper will eventually be deprecated after a migration period.
		project, projDiags := getProjectFromCredentialsFramework(ctx, client)
		if projDiags != nil {
			if !projDiags.HasError() {
				resp.Diagnostics.Append(projDiags...)
			} else {
				resp.Diagnostics.AddError("unable to get project from credentials", "")
				return
			}
		}

		client.Config.OrganizationID = project.Parent.ID
		client.Config.ProjectID = project.ID
	}

	var config ProviderFrameworkConfiguration
	config.Client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}
