// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	grpcstatus "google.golang.org/grpc/status"
)

var (
	testAccVaultPluginConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id            = "%s"
	cloud_provider    = "aws"
	region            = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
	cluster_id         = "%s"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "DEV"
}

resource "hcp_vault_plugin" "venafi_plugin" {
	cluster_id         = hcp_vault_cluster.test.cluster_id
	plugin_name        = "venafi-pki-backend"
	plugin_type        = "SECRET"
}
`, addTimestampSuffix("test-hvn-aws-"), addTimestampSuffix("test-cluster-"))

	testAccVaultPluginDataSourceConfig = fmt.Sprintf(`%s
	data "hcp_vault_plugin" "test" {
		cluster_id         = hcp_vault_cluster.test.cluster_id
		plugin_name        = "venafi-pki-backend"
		plugin_type        = "SECRET"
	}
`, testAccVaultPluginConfig)
)

func TestAccVaultPlugin(t *testing.T) {
	resourceName := "hcp_vault_plugin.venafi_plugin"
	dataSourceName := "data.hcp_vault_plugin.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVaultPluginDestroy,

		Steps: []resource.TestStep{
			// Testing Create
			{
				Config: testConfig(testAccVaultPluginConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccChecVaultPluginExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "plugin_name", "venafi-pki-backend"),
					resource.TestCheckResourceAttr(resourceName, "plugin_type", "SECRET"),
				),
			},
			// Testing that we can import Vault plugin created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					return fmt.Sprintf("%s:%s:%s:%s",
						rs.Primary.Attributes["project_id"],
						rs.Primary.Attributes["cluster_id"],
						rs.Primary.Attributes["plugin_type"],
						rs.Primary.Attributes["plugin_name"]), nil
				},
				ImportStateVerify: true,
			},
			// Testing Read
			{
				Config: testConfig(testAccVaultPluginConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccChecVaultPluginExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "plugin_name", "venafi-pki-backend"),
					resource.TestCheckResourceAttr(resourceName, "plugin_type", "SECRET"),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccVaultPluginDataSourceConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "plugin_type", dataSourceName, "plugin_type"),
					resource.TestCheckResourceAttrPair(resourceName, "plugin_name", dataSourceName, "plugin_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
				),
			},
		},
	})
}

func testAccChecVaultPluginExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*clients.Client)

		isRegistered, err := isPluginRegistered(client, id)
		if err != nil {
			return err
		}

		if !isRegistered {
			return fmt.Errorf("unable to find plugin: %q", id)
		}

		return nil
	}
}

func testAccCheckVaultPluginDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_vault_plugin":
			id := rs.Primary.ID
			isRegistered, err := isPluginRegistered(client, id)
			if err != nil {
				return err
			}
			if isRegistered {
				return fmt.Errorf("plugin status is still reporting that plugin is registered: %s", id)
			}
		default:
			continue
		}
	}
	return nil
}

func isPluginRegistered(client *clients.Client, id string) (bool, error) {
	idParts := strings.SplitN(id, "/", 8)

	clusterID := idParts[4]
	pluginType := vaultmodels.HashicorpCloudVault20201125PluginType(idParts[6])
	pluginName := idParts[7]

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	pluginsResp, err := clients.ListPlugins(context.Background(), client, loc, clusterID)
	if err != nil {
		// if cluster is deleted, plugin doesn't exist
		if clients.IsResponseCodeNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("unable to list plugins %q: %v. code: %d", id, err, grpcstatus.Code(err))
	}

	for _, plugin := range pluginsResp.Plugins {
		if strings.EqualFold(pluginName, plugin.PluginName) && pluginType == *plugin.PluginType && plugin.IsRegistered {
			return true, nil
		}
	}

	return false, nil
}
