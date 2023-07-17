// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPackerChannel(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, acctestAlpineBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() { upsertBucket(t, acctestAlpineBucket) },
				Config:    testConfig(testAccPackerChannelBasic(acctestAlpineBucket, acctestProductionChannel)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", acctestAlpineBucket),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", acctestProductionChannel),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// Testing that we can import bucket channel created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					bucketName := rs.Primary.Attributes["bucket_name"]
					channelName := rs.Primary.Attributes["name"]
					return fmt.Sprintf("%s:%s", bucketName, channelName), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

var testAccPackerChannelBasic = func(bucketName, channelName string) string {
	return fmt.Sprintf(`
	resource "hcp_packer_channel" "production" {
		bucket_name  = %q
		name = %q
	}`, bucketName, channelName)
}
