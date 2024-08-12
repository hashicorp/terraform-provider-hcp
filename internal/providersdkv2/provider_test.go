// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

// testAccProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProvider *schema.Provider

// testAccProviderConfigure ensures testAccProvider is only configured once
//
// The testAccPreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// testAccProvider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

var testProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hcp": func() (tfprotov6.ProviderServer, error) {
		// Upgrade the provider sdkv2 version to protocol 6
		upgradedSdkProvider, err := tf5to6server.UpgradeServer(
			context.Background(),
			New()().GRPCProvider,
		)
		if err != nil {
			return nil, err
		}

		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(provider.NewFrameworkProvider(version.ProviderVersion)()),
			func() tfprotov6.ProviderServer {
				return upgradedSdkProvider
			},
		}
		return tf6muxserver.NewMuxServer(context.Background(), providers...)
	},
	"dummy": func() (tfprotov6.ProviderServer, error) {
		// Upgrade the provider sdkv2 version to protocol 6
		upgradedSdkProvider, err := tf5to6server.UpgradeServer(
			context.Background(),
			testAccNewDummyProvider().GRPCProvider,
		)
		if err != nil {
			return nil, err
		}

		providers := []func() tfprotov6.ProviderServer{
			func() tfprotov6.ProviderServer {
				return upgradedSdkProvider
			},
		}
		return tf6muxserver.NewMuxServer(context.Background(), providers...)

	},
}

func init() {
	testAccProvider = New()()
}

func TestProvider(t *testing.T) {
	if err := New()().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// testAccPreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It ensures
// testing functions that attempt to call HCP APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func testAccPreCheck(t *testing.T, requiredCreds map[string]bool) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		if os.Getenv("HCP_CLIENT_ID") == "" {
			t.Fatal("HCP_CLIENT_ID must be set for acceptance tests")
		}

		if os.Getenv("HCP_CLIENT_SECRET") == "" {
			t.Fatal("HCP_CLIENT_SECRET must be set for acceptance tests")
		}

		if requiredCreds["aws"] {
			if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
				t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
			}

			if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
				t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
			}

			if os.Getenv("AWS_SESSION_TOKEN") == "" {
				t.Fatal("AWS_SESSION_TOKEN must be set for acceptance tests")
			}
		}

		if requiredCreds["azure"] {
			if os.Getenv("ARM_TENANT_ID") == "" {
				t.Fatal("ARM_TENANT_ID must be set for acceptance tests")
			}

			if os.Getenv("ARM_SUBSCRIPTION_ID") == "" {
				t.Fatal("ARM_SUBSCRIPTION_ID must be set for acceptance tests")
			}
		}

		testDiag := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if testDiag.HasError() {
			t.Fatalf("unexpected error, exiting test: %#v", testDiag)
		}
	})
}

func testConfig(res ...string) string {
	provider := `provider "hcp" {}`

	c := []string{provider}
	c = append(c, res...)
	return strings.Join(c, "\n")
}

func Test_readWorkloadIdentity(t *testing.T) {
	tcs := map[string]struct {
		config    interface{}
		want      clients.ClientConfig
		wantDiags diag.Diagnostics
	}{
		"missing token and file": {
			config: []interface{}{
				map[string]interface{}{
					"resource_name": "my_resource",
				},
			},
			wantDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "invalid workload_identity",
					Detail:   "at least one of `token_file` or `token` must be set",
				},
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
			},
		},
		"token": {
			config: []interface{}{
				map[string]interface{}{
					"resource_name": "my_resource",
					"token":         "my_token",
				},
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityToken:        "my_token",
			},
		},
		"file": {
			config: []interface{}{
				map[string]interface{}{
					"resource_name": "my_resource",
					"token_file":    "/path/to/token/file",
				},
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
			},
		},
		"both": {
			config: []interface{}{
				map[string]interface{}{
					"resource_name": "my_resource",
					"token":         "my_token",
					"token_file":    "/path/to/token/file",
				},
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityToken:        "my_token",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			// cty.Path objects are difficult to compare automatically as they
			// contain unexported fields. We can ignore them for the purposes of
			// this test.
			ignoreAttributePath := cmpopts.IgnoreFields(diag.Diagnostic{}, "AttributePath")

			got, gotDiags := readWorkloadIdentity(tc.config, clients.ClientConfig{})
			if diff := cmp.Diff(tc.wantDiags, gotDiags, ignoreAttributePath); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
