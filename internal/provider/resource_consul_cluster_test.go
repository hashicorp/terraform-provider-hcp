package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	testAccConsulClusterConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}
  
resource "hcp_consul_cluster" "test" {
	cluster_id = "test-consul-cluster"
	hvn_id     = hcp_hvn.test.hvn_id
	tier       = "development"
}`)
)

func TestAccConsulCluster(t *testing.T) {
	resourceName := "hcp_consul_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckConsulClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConsulClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-consul-cluster"),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(resourceName, "tier", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "datacenter", "test-consul-cluster"),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_interval", "24h"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_retention", "30d"),
					resource.TestCheckResourceAttr(resourceName, "connect_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_config_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_version"),
					resource.TestCheckNoResourceAttr(resourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_private_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "self_link"),
					resource.TestCheckNoResourceAttr(resourceName, "primary_link"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_accessor_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_secret_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					return rs.Primary.Attributes["cluster_id"], nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"consul_root_token_accessor_id", "consul_root_token_secret_id"},
			},
			{
				Config: testConfig(testAccConsulClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-consul-cluster"),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(resourceName, "tier", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "datacenter", "test-consul-cluster"),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_interval", "24h"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_retention", "30d"),
					resource.TestCheckResourceAttr(resourceName, "connect_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_config_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_version"),
					resource.TestCheckNoResourceAttr(resourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_private_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "self_link"),
					resource.TestCheckNoResourceAttr(resourceName, "primary_link"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_accessor_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_secret_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
				),
			},
		},
	})
}

func testAccCheckConsulClusterExists(name string) resource.TestCheckFunc {
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

		link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		clusterID := link.ID
		loc := link.Location

		if _, err := clients.GetConsulClusterByID(context.Background(), client, loc, clusterID); err != nil {
			return fmt.Errorf("unable to read Consul cluster %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckConsulClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_consul_cluster":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			clusterID := link.ID
			loc := link.Location

			_, err = clients.GetConsulClusterByID(context.Background(), client, loc, clusterID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Consul cluster %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
