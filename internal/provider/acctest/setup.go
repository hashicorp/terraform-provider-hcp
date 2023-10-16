// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"os"
	"testing"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

// ProtoV5ProviderFactories provides a Provider Factory to be used within
// acceptance tests.
var ProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"hcp": func() (tfprotov5.ProviderServer, error) {
		providers := []func() tfprotov5.ProviderServer{
			providerserver.NewProtocol5(provider.NewFrameworkProvider(version.ProviderVersion)()),
		}

		return tf5muxserver.NewMuxServer(context.Background(), providers...)
	},
}

// PreCheck verifies that the required provider testing configuration is set.
//
// This PreCheck function should be present in every acceptance test. It ensures
// credentials and other test environment settings are configured.
func PreCheck(t *testing.T) {
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

// HCPClients returns the clients from the test provider.
func HCPClients(t *testing.T) *clients.Client {
	p := provider.NewFrameworkProvider(version.ProviderVersion)()
	var resp tfprovider.ConfigureResponse
	p.Configure(context.Background(), tfprovider.ConfigureRequest{
		TerraformVersion: "",
		Config: tfsdk.Config{
			Raw:    tftypes.Value{},
			Schema: &schema.Schema{},
		},
	}, &resp)

	client, ok := resp.ResourceData.(*clients.Client)
	if !ok {
		t.Fatal("configure didn't return expected clients")
	}

	return client
}
