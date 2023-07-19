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
	bucketSlug := testAccCreateSlug("ChannelSimple")
	channelSlug := bucketSlug // No need for a different slug
	channelConfig := testAccPackerChannelBuilderBase("SimpleChannel", fmt.Sprintf("%q", channelSlug), fmt.Sprintf("%q", bucketSlug))
	unrestrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "false")
	restrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "true")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
		},
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(channelConfig)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(channelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(channelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "restricted"),
				),
			},
			{
				ResourceName:      channelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Unrestrict channel (likely a no-op)
				Config: testConfig(testAccConfigBuildersToString(unrestrictedChannelConfig)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "restricted", "false"),
				),
			},
			{ // Validate importing explicitly unrestricted channel
				ResourceName:      unrestrictedChannelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Restrict channel
				Config: testConfig(testAccConfigBuildersToString(restrictedChannelConfig)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "restricted", "true"),
				),
			},
			{ // Validate importing explicitly restricted channel
				ResourceName:      restrictedChannelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPackerChannel_HCPManaged(t *testing.T) {
	bucketSlug := testAccCreateSlug("ChannelHCPManaged")
	channelSlug := "latest"
	channelConfig := testAccPackerChannelBuilderBase("latest", fmt.Sprintf("%q", channelSlug), fmt.Sprintf("%q", bucketSlug))
	unrestrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "false")
	restrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "true")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
		},
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Validate "creating" (automatically adopting) a managed channel
				Config: testConfig(testAccConfigBuildersToString(channelConfig)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(channelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(channelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttrSet(channelConfig.ResourceName(), "restricted"),
				),
			},
			{
				ResourceName:      channelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Unrestrict managed channel
				Config: testConfig(testAccConfigBuildersToString(unrestrictedChannelConfig)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(unrestrictedChannelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttr(unrestrictedChannelConfig.ResourceName(), "restricted", "false"),
				),
			},
			{ // Validate importing explicitly unrestricted managed channel
				ResourceName:      unrestrictedChannelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Restrict managed channel
				Config: testConfig(testAccConfigBuildersToString(restrictedChannelConfig)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "author_id"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "bucket_name", bucketSlug),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "created_at"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "id"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "name", channelSlug),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "organization_id"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "project_id"),
					resource.TestCheckResourceAttrSet(restrictedChannelConfig.ResourceName(), "updated_at"),
					resource.TestCheckResourceAttr(restrictedChannelConfig.ResourceName(), "restricted", "true"),
				),
			},
			{ // Validate importing explicitly restricted managed channel
				ResourceName:      restrictedChannelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPackerChannelBuilderBase(uniqueName string, channelName string, bucketName string) testAccConfigBuilderInterface {
	return testAccPackerChannelBuilder(uniqueName, channelName, bucketName, "")
}

func testAccPackerChannelBuilderFromChannel(oldChannel testAccConfigBuilderInterface, restricted string) testAccConfigBuilderInterface {
	return testAccPackerChannelBuilder(
		oldChannel.UniqueName(),
		oldChannel.Attributes()["name"],
		oldChannel.Attributes()["bucket_name"],
		restricted,
	)
}

func testAccPackerChannelBuilder(uniqueName string, channelName string, bucketName string, restricted string) testAccConfigBuilderInterface {
	return &testAccConfigBuilder{
		resourceType: "hcp_packer_channel",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"name":        channelName,
			"bucket_name": bucketName,
			"restricted":  restricted,
		},
	}
}
