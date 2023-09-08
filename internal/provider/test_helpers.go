package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

var TestProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"hcp": func() (tfprotov5.ProviderServer, error) {
		providers := []func() tfprotov5.ProviderServer{
			providerserver.NewProtocol5(NewFrameworkProvider(version.ProviderVersion)()),
		}
		return tf5muxserver.NewMuxServer(context.Background(), providers...)
	},
}

// testPreCheck verifies and sets required provider testing configuration
//
// This testPreCheck function should be present in every acceptance test. It ensures
// credentials and other test environment settings are configured.
func testPreCheck(t *testing.T) {

	if os.Getenv("HCP_CLIENT_ID") == "" {
		t.Fatal("HCP_CLIENT_ID must be set for acceptance tests")
	}

	if os.Getenv("HCP_CLIENT_SECRET") == "" {
		t.Fatal("HCP_CLIENT_SECRET must be set for acceptance tests")
	}

	if os.Getenv("HCP_PROJECT_ID") == "" {
		t.Fatal("HCP_PROJECT_ID must be set for acceptance tests")
	}
}

// Configures a new HCP client. This allows us to not rely on the
// terraform configure method for test setup
func NewClient() (*clients.Client, error) {
	client, err := clients.NewClient(clients.ClientConfig{
		ClientID:      os.Getenv("HCP_CLIENT_ID"),
		ClientSecret:  os.Getenv("HCP_CLIENT_SECRET"),
		SourceChannel: "terraform-provider-hcp",
	})
	if err != nil {
		return nil, err
	}
	projectID := os.Getenv("HCP_PROJECT_ID")

	getProjParams := project_service.NewProjectServiceGetParams()
	getProjParams.ID = projectID
	project, err := clients.RetryProjectServiceGet(client, getProjParams)
	if err != nil {
		return nil, err
	}

	client.Config.ProjectID = project.Payload.Project.ID
	client.Config.OrganizationID = project.Payload.Project.Parent.ID
	if err != nil {
		return nil, err
	}
	return client, nil
}
