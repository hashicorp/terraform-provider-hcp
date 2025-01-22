// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logstreaming_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccHCPLogStreamingDestinationSplunk(t *testing.T) {
	resourceName := "hcp_log_streaming_destination.test_splunk_cloud"
	spName := "splunk-resource-name-1"
	spNameUpdated := "splunk-resource-name-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			err := testAccHCPLogStreamingDestinationDestroy(t, s)
			if err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testAccSplunkConfig(spName),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", spName),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.token"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.endpoint", "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.token", "splunk-authentication-token"),
				),
			},
			{
				// Update the name and token and expect in-place update
				Config: testAccSplunkConfigUpdated(spNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", spNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.token"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.endpoint", "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.token", "splunk-authentication-token234"),
				),
			},
		},
	})
}

func testAccSplunkConfig(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_splunk_cloud" {
  			name = "%[1]s"
  			splunk_cloud = {
  				endpoint = "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"
  				token = "splunk-authentication-token"
  			}
  		}
  		`, name)
}

func testAccSplunkConfigUpdated(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_splunk_cloud" {
  			name = "%[1]s"
  			splunk_cloud = {
  				endpoint = "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"
  				token = "splunk-authentication-token234"
  			}
  		}
  		`, name)
}

func TestAccHCPLogStreamingDestinationCloudWatch(t *testing.T) {
	resourceName := "hcp_log_streaming_destination.test_cloudwatch"
	cwName := "cloudwatch-resource-name-1"
	cwNameUpdated := "cloudwatch-resource-name-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			err := testAccHCPLogStreamingDestinationDestroy(t, s)
			if err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testAccCloudWatchLogsConfig(cwName),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", cwName),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.region"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.external_id"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.external_id", "superSecretExternalID"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.role_arn", "arn:aws:iam::000000000000:role/cloud_watch_role"),
				),
			},
			{
				// Update the name, log group name and externalID and expect in-place update
				Config: testAccCloudWatchLogsConfigUpdated(cwNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", cwNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.region"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.log_group_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.external_id", "superSecretExternalID789"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.role_arn", "arn:aws:iam::000000000000:role/cloud_watch_role"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.log_group_name", "a-log-group-name"),
				),
			},
		},
	})
}

func testAccCloudWatchLogsConfig(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_cloudwatch" {
  			name = "%[1]s"
  			cloudwatch = {
  				region = "us-west-2"
  				role_arn = "arn:aws:iam::000000000000:role/cloud_watch_role"
				external_id = "superSecretExternalID"
  			}
  		}
  		`, name)
}

func testAccCloudWatchLogsConfigUpdated(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_cloudwatch" {
  			name = "%[1]s"
  			cloudwatch = {
  				region = "us-west-2"
  				role_arn = "arn:aws:iam::000000000000:role/cloud_watch_role"
				external_id = "superSecretExternalID789"
				log_group_name = "a-log-group-name"
  			}
  		}
  		`, name)
}

func TestAccHCPLogStreamingDestinationDatadog(t *testing.T) {
	resourceName := "hcp_log_streaming_destination.test_datadog"
	ddName := "dd-resource-name-1"
	ddNameUpdated := "dd-resource-name-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			err := testAccHCPLogStreamingDestinationDestroy(t, s)
			if err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testAccDatadogConfig(ddName),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", ddName),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.application_key"),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.api_key"),
					resource.TestCheckResourceAttr(resourceName, "datadog.endpoint", "https://datadog-api.com"),
					resource.TestCheckResourceAttr(resourceName, "datadog.application_key", "APPLICATION-VALUE-HERE"),
					resource.TestCheckResourceAttr(resourceName, "datadog.api_key", "VALUEHERE"),
				),
			},
			{
				// Update the name, endpoint and api key and expect in-place update
				Config: testAccDatadogConfigUpdated(ddNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", ddNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.application_key"),
					resource.TestCheckResourceAttrSet(resourceName, "datadog.api_key"),
					resource.TestCheckResourceAttr(resourceName, "datadog.endpoint", "https://datadog-api.com/updated-endpoint"),
					resource.TestCheckResourceAttr(resourceName, "datadog.application_key", "APPLICATION-VALUE-HERE"),
					resource.TestCheckResourceAttr(resourceName, "datadog.api_key", "VALUEHERECHANGED"),
				),
			},
		},
	})
}

func testAccDatadogConfig(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_datadog" {
  			name = "%[1]s"
  			datadog = {
  				endpoint = "https://datadog-api.com"
				api_key = "VALUEHERE"
				application_key = "APPLICATION-VALUE-HERE"
  			}
  		}
  		`, name)
}

func testAccDatadogConfigUpdated(name string) string {
	return fmt.Sprintf(`
 		resource "hcp_log_streaming_destination" "test_datadog" {
 			name = "%[1]s"
 			datadog = {
 				endpoint = "https://datadog-api.com/updated-endpoint"
				api_key = "VALUEHERECHANGED"
				application_key = "APPLICATION-VALUE-HERE"
 			}
 		}
 		`, name)
}

func testAccHCPLogStreamingDestinationExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		streamingDestinationID := rs.Primary.Attributes["streaming_destination_id"]
		if streamingDestinationID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := acctest.HCPClients(t)

		res, err := clients.GetLogStreamingDestination(context.Background(), client, client.Config.OrganizationID, streamingDestinationID)
		if err != nil {
			return fmt.Errorf("unable to read streaming destination %q: %v", streamingDestinationID, err)
		}

		if res == nil {
			return fmt.Errorf("log Streaming Destination (%s) not found", streamingDestinationID)
		}

		return nil
	}
}

func testAccHCPLogStreamingDestinationDestroy(t *testing.T, s *terraform.State) error {
	client := acctest.HCPClients(t)
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_log_streaming_destination":

			streamingDestinationID := rs.Primary.Attributes["streaming_destination_id"]

			_, err := clients.GetLogStreamingDestination(context.Background(), client, client.Config.OrganizationID, streamingDestinationID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed streaming destination %q: %v", streamingDestinationID, err)
			}

		default:
			continue
		}
	}
	return nil
}
