package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type inputT struct {
	vaultClusterName           string
	hvnName                    string
	vaultClusterResourceName   string
	vaultClusterDataSourceName string
	adminTokenResourceName     string
	tf                         string
	cloudProvider              string
	region                     string
	tier                       string
	updateTier1                string
	updateTier2                string
	publicEndpoint             string
}

// This includes tests against both the resource, the corresponding datasource, and the dependent admin token resource
// to shorten testing time.
func TestAccVaultClusterAzure(t *testing.T) {
	azureTestInput := inputT{
		vaultClusterName:           "test-vault-cluster-azure",
		hvnName:                    "test-hvn-azure",
		vaultClusterResourceName:   vaultClusterResourceName,
		vaultClusterDataSourceName: vaultClusterDataSourceName,
		adminTokenResourceName:     adminTokenResourceName,
		cloudProvider:              cloudProviderAzure,
		region:                     azureRegion,
		tier:                       "DEV",
	}
	tf := setTestAccVaultClusterConfig(t, vaultCluster, azureTestInput, azureTestInput.tier)
	// save so e don't have to generate this again and again
	azureTestInput.tf = tf
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps:             azureTestSteps(t, azureTestInput),
	})
}

// This includes tests against both the resource, the corresponding datasource, and the dependent admin token resource
// to shorten testing time.
func TestAccVaultClusterAWS(t *testing.T) {
	awsTestInput := inputT{
		vaultClusterName:           "test-vault-cluster-aws",
		hvnName:                    "test-hvn-aws",
		vaultClusterResourceName:   vaultClusterResourceName,
		vaultClusterDataSourceName: vaultClusterDataSourceName,
		adminTokenResourceName:     adminTokenResourceName,
		cloudProvider:              cloudProviderAWS,
		region:                     awsRegion,
		tier:                       "DEV",
		updateTier1:                "STANDARD_SMALL",
		updateTier2:                "STANDARD_MEDIUM",
		publicEndpoint:             "false",
	}

	tf := setTestAccVaultClusterConfig(t, vaultCluster, awsTestInput, awsTestInput.tier)
	// save so e don't have to generate this again and again
	awsTestInput.tf = tf
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps:             awsTestSteps(t, awsTestInput),
	})
}

func testAccCheckVaultClusterExists(name string) resource.TestCheckFunc {
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

		link, err := buildLinkFromURL(id, VaultClusterResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		clusterID := link.ID
		loc := link.Location

		if _, err := clients.GetVaultClusterByID(context.Background(), client, loc, clusterID); err != nil {
			return fmt.Errorf("unable to read Vault cluster %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckVaultClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_vault_cluster":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, VaultClusterResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			clusterID := link.ID
			loc := link.Location

			_, err = clients.GetVaultClusterByID(context.Background(), client, loc, clusterID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Vault cluster %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}

// utlity functions
func awsTestSteps(t *testing.T, inp inputT) []resource.TestStep {
	in := &inp
	return []resource.TestStep{
		createClusteAndTestAdminTokenGeneration(t, in),
		importResourcesInTFState(t, in),
		tfApply(t, in),
		testTFDataSources(t, in),
		updateClusterTier(t, in),
		updateVaultPublicEndpointObservabilityDataAndMVU(t, in),
		updateTierPublicEndpointAndRemoveObservabilityData(t, in),
	}
}

func azureTestSteps(t *testing.T, inp inputT) []resource.TestStep {
	in := &inp
	return []resource.TestStep{
		createClusteAndTestAdminTokenGeneration(t, in),
		importResourcesInTFState(t, in),
		tfApply(t, in),
		testTFDataSources(t, in),
	}
}

// This step tests Vault cluster and admin token resource creation.
func createClusteAndTestAdminTokenGeneration(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		Config: testConfig(in.tf),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.vaultClusterResourceName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "cluster_id", in.vaultClusterName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "hvn_id", in.hvnName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "tier", in.tier),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "cloud_provider", in.cloudProvider),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "region", in.region),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "namespace", "admin"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_version"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "organization_id"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "project_id"),
			resource.TestCheckNoResourceAttr(in.vaultClusterResourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_private_endpoint_url", ""),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "state"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "created_at"),

			// Verifies admin token
			resource.TestCheckResourceAttr(in.adminTokenResourceName, "cluster_id", in.vaultClusterName),
			resource.TestCheckResourceAttrSet(in.adminTokenResourceName, "token"),
			resource.TestCheckResourceAttrSet(in.adminTokenResourceName, "created_at"),
		),
	}
}

// This step simulates an import of the resource
func importResourcesInTFState(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		ResourceName: in.vaultClusterResourceName,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			rs, ok := s.RootModule().Resources[in.vaultClusterResourceName]
			if !ok {
				return "", fmt.Errorf("not found: %s", in.vaultClusterResourceName)
			}

			return rs.Primary.Attributes["cluster_id"], nil
		},
		ImportStateVerify: true,
	}
}

// This step is a subsequent terraform apply that verifies that no state is modified
func tfApply(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		Config: testConfig(in.tf),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.vaultClusterResourceName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "cluster_id", in.vaultClusterName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "hvn_id", in.hvnName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "tier", in.tier),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "cloud_provider", in.cloudProvider),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "region", in.region),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "namespace", "admin"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "organization_id"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "project_id"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_version"),
			resource.TestCheckNoResourceAttr(in.vaultClusterResourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "state"),
		),
	}
}

// Tests datasource
func testTFDataSources(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		Config: testConfig(in.tf),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "cluster_id", in.vaultClusterDataSourceName, "cluster_id"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "hvn_id", in.vaultClusterDataSourceName, "hvn_id"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "public_endpoint", in.vaultClusterDataSourceName, "public_endpoint"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "min_vault_version", in.vaultClusterDataSourceName, "min_vault_version"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "tier", in.vaultClusterDataSourceName, "tier"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "organization_id", in.vaultClusterDataSourceName, "organization_id"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "project_id", in.vaultClusterDataSourceName, "project_id"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "cloud_provider", in.vaultClusterDataSourceName, "cloud_provider"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "region", in.vaultClusterDataSourceName, "region"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "namespace", in.vaultClusterDataSourceName, "namespace"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "vault_version", in.vaultClusterDataSourceName, "vault_version"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "vault_public_endpoint_url", in.vaultClusterDataSourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "vault_private_endpoint_url", in.vaultClusterDataSourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "created_at", in.vaultClusterDataSourceName, "created_at"),
			resource.TestCheckResourceAttrPair(in.vaultClusterResourceName, "state", in.vaultClusterDataSourceName, "state"),
		),
	}
}

// This step verifies the successful update of "tier" and MVU config
func updateClusterTier(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterTierAndMVUConfig, newIn, newIn.updateTier1)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.vaultClusterResourceName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "tier", in.updateTier1),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "SCHEDULED"),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_day", "WEDNESDAY"),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_time", "WINDOW_12AM_4AM"),
		),
	}
}

// This step verifies the successful update of "public_endpoint", "audit_log", "metrics" and MVU config
func updateVaultPublicEndpointObservabilityDataAndMVU(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	newIn.publicEndpoint = "true"
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterPublicAndMetricsAuditLog, newIn, newIn.updateTier1)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.vaultClusterResourceName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "public_endpoint", "true"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_public_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_public_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "metrics_config.0.splunk_hecendpoint", "https://http-input-splunkcloud.com"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "metrics_config.0.splunk_token"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "audit_log_config.0.datadog_api_key"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "audit_log_config.0.datadog_region", "us1"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "MANUAL"),
		),
	}
}

// This step verifies the successful update of both "tier" and "public_endpoint" and removal of "metrics" and "audit_log"
func updateTierPublicEndpointAndRemoveObservabilityData(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	newIn.publicEndpoint = "false"
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterTierAndMVUConfig, newIn, newIn.updateTier2)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.vaultClusterResourceName),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "tier", in.updateTier2),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_public_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_public_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.vaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckNoResourceAttr(in.vaultClusterResourceName, "metrics_config.0"),
			resource.TestCheckNoResourceAttr(in.vaultClusterResourceName, "audit_log_config.0"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "SCHEDULED"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_day", "WEDNESDAY"),
			resource.TestCheckResourceAttr(in.vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_time", "WINDOW_12AM_4AM"),
		),
	}
}
