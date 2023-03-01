// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"hcp": func() (*schema.Provider, error) {
		return New()(), nil
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

		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			if err[0].Severity == 1 {
				// Severity 1 = Warning, which can be ignored during test
				return
			}

			t.Fatalf("unexpected error, exiting test: %#v", err)
		}
	})
}

func testConfig(res ...string) string {
	provider := `provider "hcp" {}`

	c := []string{provider}
	c = append(c, res...)
	return strings.Join(c, "\n")
}
