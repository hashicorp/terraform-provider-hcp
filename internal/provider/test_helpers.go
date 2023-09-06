package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
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

// testAccPreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It ensures
// testing functions that attempt to call HCP APIs are previously configured.
func TestAccPreCheck(t *testing.T) {

	if os.Getenv("HCP_CLIENT_ID") == "" {
		t.Fatal("HCP_CLIENT_ID must be set for acceptance tests")
	}

	if os.Getenv("HCP_CLIENT_SECRET") == "" {
		t.Fatal("HCP_CLIENT_SECRET must be set for acceptance tests")
	}
}
