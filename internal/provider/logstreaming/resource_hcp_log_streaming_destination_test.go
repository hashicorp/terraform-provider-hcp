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

// This includes tests against both the resource and the corresponding datasource
// to shorten testing time.
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
