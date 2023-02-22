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

func TestAccPackerChannel_AssignedIteration(t *testing.T) {
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
				PreConfig: func() {
					fingerprint := "channel-assigned-iteration"
					upsertBucket(t, acctestAlpineBucket)
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
				},
				Config: testConfig(testAccPackerChannelAssignedLatestIteration(acctestAlpineBucket, acctestProductionChannel)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", acctestAlpineBucket),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-assigned-iteration"),
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

func TestAccPackerChannel_UpdateAssignedIteration(t *testing.T) {
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
				PreConfig: func() {
					fingerprint := "channel-update-it1"
					upsertBucket(t, acctestAlpineBucket)
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
				},
				Config: testConfig(testAccPackerChannelAssignedLatestIteration(acctestAlpineBucket, acctestProductionChannel)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", acctestAlpineBucket),
					resource.TestCheckResourceAttr(resourceName, "name", acctestProductionChannel),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-update-it1"),
				),
			},
			{
				PreConfig: func() {
					fingerprint := "channel-update-it2"
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
				},
				Config: testConfig(testAccPackerChannelAssignedLatestIteration(acctestAlpineBucket, acctestProductionChannel)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", acctestAlpineBucket),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "iteration.0.fingerprint", "channel-update-it2"),
					resource.TestCheckResourceAttr(resourceName, "name", acctestProductionChannel),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccPackerChannel_UpdateAssignedIterationWithFingerprint(t *testing.T) {
	resourceName := "hcp_packer_channel.production"

	fingerprint := "channel-update-it1"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, acctestAlpineBucket, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertBucket(t, acctestAlpineBucket)
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
				},
				Config: testConfig(testAccPackerChannelIterationFingerprint(acctestAlpineBucket, acctestProductionChannel, fingerprint)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "author_id"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", acctestAlpineBucket),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "iteration.0.incremental_version"),
					resource.TestCheckResourceAttr(resourceName, "name", acctestProductionChannel),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
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

var testAccPackerChannelAssignedLatestIteration = func(bucketName, channelName string) string {
	return fmt.Sprintf(`
	data "hcp_packer_image_iteration" "test" {
		bucket_name = %[2]q
		channel     = "latest"
	}
	resource "hcp_packer_channel" "production" {
		name = %[1]q
		bucket_name  = %[2]q
		iteration {
		  id = data.hcp_packer_image_iteration.test.id
		}
	}`, channelName, bucketName)
}

var testAccPackerChannelIterationFingerprint = func(bucketName, channelName, fingerprint string) string {
	return fmt.Sprintf(`
	resource "hcp_packer_channel" "production" {
		bucket_name  = %q
		name = %q
		iteration {
		  fingerprint = %q
		}
	}`, bucketName, channelName, fingerprint)
}
