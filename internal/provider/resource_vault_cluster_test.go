// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type inputT struct {
	VaultClusterName           string
	HvnName                    string
	HvnCidr                    string // optional
	VaultClusterResourceName   string
	VaultClusterDataSourceName string
	AdminTokenResourceName     string
	CloudProvider              string
	Region                     string
	Tier                       string
	UpdateTier1                string
	UpdateTier2                string
	PublicEndpoint             string
	Secondary                  *inputT // optional
	tf                         string
}

func (in *inputT) GetHvnCidr() string {
	if in.HvnCidr == "" {
		return "172.25.16.0/20"
	}

	return in.HvnCidr
}

// This includes tests against both the resource, the corresponding datasource, and the dependent admin token resource
// to shorten testing time.
func TestAccVaultClusterAzure(t *testing.T) {
	azureTestInput := inputT{
		VaultClusterName:           addTimestampSuffix("test-vault-azure-"),
		HvnName:                    addTimestampSuffix("test-hvn-azure-"),
		VaultClusterResourceName:   vaultClusterResourceName,
		VaultClusterDataSourceName: vaultClusterDataSourceName,
		AdminTokenResourceName:     adminTokenResourceName,
		CloudProvider:              cloudProviderAzure,
		Region:                     azureRegion,
		Tier:                       "DEV",
		UpdateTier1:                "STANDARD_SMALL",
		UpdateTier2:                "STANDARD_MEDIUM",
		PublicEndpoint:             "false",
	}
	tf := setTestAccVaultClusterConfig(t, vaultCluster, azureTestInput, azureTestInput.Tier)
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
		VaultClusterName:           addTimestampSuffix("test-vault-aws-"),
		HvnName:                    addTimestampSuffix("test-hvn-aws-"),
		VaultClusterResourceName:   vaultClusterResourceName,
		VaultClusterDataSourceName: vaultClusterDataSourceName,
		AdminTokenResourceName:     adminTokenResourceName,
		CloudProvider:              cloudProviderAWS,
		Region:                     awsRegion,
		Tier:                       "DEV",
		UpdateTier1:                "STANDARD_SMALL",
		UpdateTier2:                "STANDARD_MEDIUM",
		PublicEndpoint:             "false",
	}

	tf := setTestAccVaultClusterConfig(t, vaultCluster, awsTestInput, awsTestInput.Tier)
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

// utility functions
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
		updateClusterTier(t, in),
	}
}

// This step tests Vault cluster and admin token resource creation.
func createClusteAndTestAdminTokenGeneration(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		Config: testConfig(in.tf),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.VaultClusterResourceName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "cluster_id", in.VaultClusterName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "hvn_id", in.HvnName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "tier", in.Tier),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "cloud_provider", in.CloudProvider),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "region", in.Region),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "namespace", "admin"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_version"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "organization_id"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "project_id"),
			resource.TestCheckNoResourceAttr(in.VaultClusterResourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_private_endpoint_url", ""),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "state"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "created_at"),

			// Verifies admin token
			resource.TestCheckResourceAttr(in.AdminTokenResourceName, "cluster_id", in.VaultClusterName),
			resource.TestCheckResourceAttrSet(in.AdminTokenResourceName, "token"),
			resource.TestCheckResourceAttrSet(in.AdminTokenResourceName, "created_at"),
		),
	}
}

// This step simulates an import of the resource
func importResourcesInTFState(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		ResourceName: in.VaultClusterResourceName,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			rs, ok := s.RootModule().Resources[in.VaultClusterResourceName]
			if !ok {
				return "", fmt.Errorf("not found: %s", in.VaultClusterResourceName)
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
			testAccCheckVaultClusterExists(in.VaultClusterResourceName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "cluster_id", in.VaultClusterName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "hvn_id", in.HvnName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "tier", in.Tier),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "cloud_provider", in.CloudProvider),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "region", in.Region),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "namespace", "admin"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "organization_id"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "project_id"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_version"),
			resource.TestCheckNoResourceAttr(in.VaultClusterResourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "state"),
		),
	}
}

// Tests datasource
func testTFDataSources(t *testing.T, in *inputT) resource.TestStep {
	return resource.TestStep{
		Config: testConfig(in.tf),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "cluster_id", in.VaultClusterDataSourceName, "cluster_id"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "hvn_id", in.VaultClusterDataSourceName, "hvn_id"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "public_endpoint", in.VaultClusterDataSourceName, "public_endpoint"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "min_vault_version", in.VaultClusterDataSourceName, "min_vault_version"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "tier", in.VaultClusterDataSourceName, "tier"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "organization_id", in.VaultClusterDataSourceName, "organization_id"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "project_id", in.VaultClusterDataSourceName, "project_id"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "cloud_provider", in.VaultClusterDataSourceName, "cloud_provider"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "region", in.VaultClusterDataSourceName, "region"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "namespace", in.VaultClusterDataSourceName, "namespace"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "vault_version", in.VaultClusterDataSourceName, "vault_version"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "vault_public_endpoint_url", in.VaultClusterDataSourceName, "vault_public_endpoint_url"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "vault_private_endpoint_url", in.VaultClusterDataSourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "created_at", in.VaultClusterDataSourceName, "created_at"),
			resource.TestCheckResourceAttrPair(in.VaultClusterResourceName, "state", in.VaultClusterDataSourceName, "state"),
		),
	}
}

// This step verifies the successful update of "tier" and MVU config
func updateClusterTier(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterTierAndMVUConfig, newIn, newIn.UpdateTier1)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.VaultClusterResourceName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "tier", in.UpdateTier1),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "SCHEDULED"),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_day", "WEDNESDAY"),
			resource.TestCheckResourceAttr(vaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_time", "WINDOW_12AM_4AM"),
		),
	}
}

// This step verifies the successful update of "public_endpoint", "audit_log", "metrics" and MVU config
func updateVaultPublicEndpointObservabilityDataAndMVU(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	newIn.PublicEndpoint = "true"
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterPublicAndMetricsAuditLog, newIn, newIn.UpdateTier1)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.VaultClusterResourceName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "public_endpoint", "true"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_public_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_public_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "metrics_config.0.splunk_hecendpoint", "https://http-input-splunkcloud.com"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "metrics_config.0.splunk_token"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "audit_log_config.0.datadog_api_key"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "audit_log_config.0.datadog_region", "us1"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "MANUAL"),
		),
	}
}

// This step verifies the successful update of both "tier" and "public_endpoint" and removal of "metrics" and "audit_log"
func updateTierPublicEndpointAndRemoveObservabilityData(t *testing.T, in *inputT) resource.TestStep {
	newIn := *in
	newIn.PublicEndpoint = "false"
	return resource.TestStep{
		Config: testConfig(setTestAccVaultClusterConfig(t, updatedVaultClusterTierAndMVUConfig, newIn, newIn.UpdateTier2)),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckVaultClusterExists(in.VaultClusterResourceName),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "tier", in.UpdateTier2),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "public_endpoint", "false"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_public_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_public_endpoint_url", "8200"),
			resource.TestCheckResourceAttrSet(in.VaultClusterResourceName, "vault_private_endpoint_url"),
			testAccCheckFullURL(in.VaultClusterResourceName, "vault_private_endpoint_url", "8200"),
			resource.TestCheckNoResourceAttr(in.VaultClusterResourceName, "metrics_config.0"),
			resource.TestCheckNoResourceAttr(in.VaultClusterResourceName, "audit_log_config.0"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "major_version_upgrade_config.0.upgrade_type", "SCHEDULED"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_day", "WEDNESDAY"),
			resource.TestCheckResourceAttr(in.VaultClusterResourceName, "major_version_upgrade_config.0.maintenance_window_time", "WINDOW_12AM_4AM"),
		),
	}
}

func addTimestampSuffix(in string) string {
	return in + time.Now().Format("200601021504")
}
