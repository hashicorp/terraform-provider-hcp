package acctest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

// RandString generates a random string with the given length.
func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// CheckDataSourceStateMatchesResourceStateWithIgnores is a helper that allows
// comparing the state of a data source to that of a resource. For example, an
// acceptance test can uses the following config:
//
//	resource "hcp_project" "project" {
//	  name        = %q
//	  description = %q
//	}
//
//	data "hcp_project" "project" {
//	  project = hcp_project.project.resource_id
//	}
//
// Then the following check can be defined:
//
//	  CheckDataSourceStateMatchesResourceStateWithIgnores(
//		 "data.hcp_project.project",
//	     "hcp_project.project",
//	      map[string]struct{}{"project": {}}),
//	  )
func CheckDataSourceStateMatchesResourceStateWithIgnores(dataSourceName, resourceName string, ignoreFields map[string]struct{}) func(*terraform.State) error {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", dataSourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		dsAttr := ds.Primary.Attributes
		rsAttr := rs.Primary.Attributes

		errMsg := ""
		// Data sources are often derived from resources, so iterate over the resource fields to
		// make sure all fields are accounted for in the data source.
		// If a field exists in the data source but not in the resource, its expected value should
		// be checked separately.
		for k := range rsAttr {
			if _, ok := ignoreFields[k]; ok {
				continue
			}
			if _, ok := ignoreFields["labels.%"]; ok && strings.HasPrefix(k, "labels.") {
				continue
			}
			if _, ok := ignoreFields["terraform_labels.%"]; ok && strings.HasPrefix(k, "terraform_labels.") {
				continue
			}
			if k == "%" {
				continue
			}
			if dsAttr[k] != rsAttr[k] {
				// ignore data sources where an empty list is being compared against a null list.
				if k[len(k)-1:] == "#" && (dsAttr[k] == "" || dsAttr[k] == "0") && (rsAttr[k] == "" || rsAttr[k] == "0") {
					continue
				}
				errMsg += fmt.Sprintf("%s is %s; want %s\n", k, dsAttr[k], rsAttr[k])
			}
		}

		if errMsg != "" {
			return errors.New(errMsg)
		}

		return nil
	}
}
