// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logstreaming_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccHCPLogStreamingDestinationSplunk(t *testing.T) {
	resourceName := "hcp_log_streaming_destination.test_splunk_cloud"
	spName := "splunk-resource-name-1"
	spNameUpdated := "splunk-resource-name-2"
	var sp models.LogService20210330Destination
	var sp2 models.LogService20210330Destination

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
					testAccHCPLogStreamingDestinationExists(t, resourceName, &sp),
					resource.TestCheckResourceAttr(resourceName, "name", spName),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.token"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.endpoint", "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.token", "splunk-authentication-token"),
				),
			},
			{
				// Update the name
				Config: testAccSplunkConfig(spNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName, &sp2),
					resource.TestCheckResourceAttr(resourceName, "name", spNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "splunk_cloud.token"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.endpoint", "https://http-inputs-hcptest.splunkcloud.com/services/collector/event"),
					resource.TestCheckResourceAttr(resourceName, "splunk_cloud.token", "splunk-authentication-token"),
					func(_ *terraform.State) error {
						if sp.Resource.ID == sp2.Resource.ID {
							return fmt.Errorf("resource_ids match, indicating resource wasn't recreated")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccHCPLogStreamingDestinationCloudWatch(t *testing.T) {
	resourceName := "hcp_log_streaming_destination.test_cloudwatch"
	cwName := "cloudwatch-resource-name-1"
	cwNameUpdated := "cloudwatch-resource-name-2"
	cwNameLogGroup := "cloudwatch-resource-name-3"
	var cw models.LogService20210330Destination
	var cw2 models.LogService20210330Destination
	var cw3 models.LogService20210330Destination

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
					testAccHCPLogStreamingDestinationExists(t, resourceName, &cw),
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
				// Update the name
				Config: testAccCloudWatchLogsConfig(cwNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName, &cw2),
					resource.TestCheckResourceAttr(resourceName, "name", cwNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.region"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.external_id"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.external_id", "superSecretExternalID"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.role_arn", "arn:aws:iam::000000000000:role/cloud_watch_role"),
					func(_ *terraform.State) error {
						if cw.Resource.ID == cw2.Resource.ID {
							return fmt.Errorf("resource_ids match, indicating resource wasn't recreated")
						}
						return nil
					},
				),
			},
			{
				// test with a log group name
				Config: testAccCloudWatchLogsConfigLogGroupName(cwNameLogGroup),
				Check: resource.ComposeTestCheckFunc(
					testAccHCPLogStreamingDestinationExists(t, resourceName, &cw3),
					resource.TestCheckResourceAttr(resourceName, "name", cwNameLogGroup),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.region"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch.log_group_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.external_id", "superSecretExternalID"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.role_arn", "arn:aws:iam::000000000000:role/cloud_watch_role"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch.log_group_name", "a-log-group-name"),
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

func testAccCloudWatchLogsConfigLogGroupName(name string) string {
	return fmt.Sprintf(`
  		resource "hcp_log_streaming_destination" "test_cloudwatch" {
  			name = "%[1]s"
  			cloudwatch = {
  				region = "us-west-2"
  				role_arn = "arn:aws:iam::000000000000:role/cloud_watch_role"
				external_id = "superSecretExternalID"
				log_group_name = "a-log-group-name"
  			}
  		}
  		`, name)
}

func testAccHCPLogStreamingDestinationExists(t *testing.T, name string, destination *models.LogService20210330Destination) resource.TestCheckFunc {
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
		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      client.Config.ProjectID,
		}

		res, err := clients.GetLogStreamingDestination(context.Background(), client, loc, streamingDestinationID)
		if err != nil {
			return fmt.Errorf("unable to read streaming destination %q: %v", streamingDestinationID, err)
		}

		if res == nil {
			return fmt.Errorf("log Streaming Destination (%s) not found", streamingDestinationID)
		}

		// assign the response to the pointer
		*destination = *res
		return nil
	}
}

func testAccHCPLogStreamingDestinationDestroy(t *testing.T, s *terraform.State) error {
	client := acctest.HCPClients(t)
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_log_streaming_destination":

			streamingDestinationID := rs.Primary.Attributes["streaming_destination_id"]

			loc := &sharedmodels.HashicorpCloudLocationLocation{
				OrganizationID: client.Config.OrganizationID,
				ProjectID:      client.Config.ProjectID,
			}

			_, err := clients.GetLogStreamingDestination(context.Background(), client, loc, streamingDestinationID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed streaming destination %q: %v", streamingDestinationID, err)
			}

		default:
			continue
		}
	}
	return nil
}
