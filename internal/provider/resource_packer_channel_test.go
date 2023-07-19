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
				Check:  testAccCheckPackerChannel(channelConfig.ResourceName(), channelSlug, bucketSlug, ""),
			},
			{
				ResourceName:      channelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Unrestrict channel (likely a no-op)
				Config: testConfig(testAccConfigBuildersToString(unrestrictedChannelConfig)),
				Check:  testAccCheckPackerChannel(unrestrictedChannelConfig.ResourceName(), channelSlug, bucketSlug, "false"),
			},
			{ // Validate importing explicitly unrestricted channel
				ResourceName:      unrestrictedChannelConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Restrict channel
				Config: testConfig(testAccConfigBuildersToString(restrictedChannelConfig)),
				Check:  testAccCheckPackerChannel(restrictedChannelConfig.ResourceName(), channelSlug, bucketSlug, "true"),
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
	latestConfig := testAccPackerChannelBuilderBase("latest", fmt.Sprintf("%q", channelSlug), fmt.Sprintf("%q", bucketSlug))
	unrestrictedLatestConfig := testAccPackerChannelBuilderFromChannel(latestConfig, "false")
	restrictedLatestConfig := testAccPackerChannelBuilderFromChannel(latestConfig, "true")

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
				Config: testConfig(testAccConfigBuildersToString(latestConfig)),
				Check:  testAccCheckPackerChannel(latestConfig.ResourceName(), channelSlug, bucketSlug, ""),
			},
			{
				ResourceName:      latestConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Unrestrict managed channel
				Config: testConfig(testAccConfigBuildersToString(unrestrictedLatestConfig)),
				Check:  testAccCheckPackerChannel(unrestrictedLatestConfig.ResourceName(), channelSlug, bucketSlug, "false"),
			},
			{ // Validate importing explicitly unrestricted managed channel
				ResourceName:      unrestrictedLatestConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Restrict managed channel
				Config: testConfig(testAccConfigBuildersToString(restrictedLatestConfig)),
				Check:  testAccCheckPackerChannel(restrictedLatestConfig.ResourceName(), channelSlug, bucketSlug, "true"),
			},
			{ // Validate importing explicitly restricted managed channel
				ResourceName:      restrictedLatestConfig.ResourceName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPackerChannel_RestrictionDrift(t *testing.T) {
	bucketSlug := testAccCreateSlug("RestrictionDrift")
	channelSlug := bucketSlug // No need for a different slug

	channelUnrestrictedConfig := testAccPackerChannelBuilder("Drift", fmt.Sprintf("%q", channelSlug), fmt.Sprintf("%q", bucketSlug), "false")
	channelRestrictedConfig := testAccPackerChannelBuilderFromChannel(channelUnrestrictedConfig, "true")

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
			// Normal channels
			{
				Config: testConfig(testAccConfigBuildersToString(channelUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelUnrestrictedConfig.ResourceName(), channelSlug, bucketSlug, "false"),
			},
			{ // Check drift mitigation for normal channel from false->true
				PreConfig: func() {
					updateChannelRestriction(t, bucketSlug, channelSlug, true)
				},
				Config: testConfig(testAccConfigBuildersToString(channelUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelUnrestrictedConfig.ResourceName(), channelSlug, bucketSlug, "false"),
			},
			{
				Config: testConfig(testAccConfigBuildersToString(channelRestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelRestrictedConfig.ResourceName(), channelSlug, bucketSlug, "true"),
			},
			{ // Check drift mitigation for normal channel from true->false
				PreConfig: func() {
					updateChannelRestriction(t, bucketSlug, channelSlug, false)
				},
				Config: testConfig(testAccConfigBuildersToString(channelRestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelRestrictedConfig.ResourceName(), channelSlug, bucketSlug, "true"),
			},
		},
	})
}

func TestAccPackerChannel_RestrictionDriftHCPManaged(t *testing.T) {
	bucketSlug := testAccCreateSlug("RestrictionDriftHCPManaged")
	latestSlug := "latest"

	latestUnrestrictedConfig := testAccPackerChannelBuilder("DriftHCPManaged", fmt.Sprintf("%q", latestSlug), fmt.Sprintf("%q", bucketSlug), "false")
	latestRestrictedConfig := testAccPackerChannelBuilderFromChannel(latestUnrestrictedConfig, "true")

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
				Config: testConfig(testAccConfigBuildersToString(latestUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestUnrestrictedConfig.ResourceName(), latestSlug, bucketSlug, "false"),
			},
			{ // Check drift mitigation for HCP managed channel from false->true
				PreConfig: func() {
					updateChannelRestriction(t, bucketSlug, latestSlug, true)
				},
				Config: testConfig(testAccConfigBuildersToString(latestUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestUnrestrictedConfig.ResourceName(), latestSlug, bucketSlug, "false"),
			},
			{
				Config: testConfig(testAccConfigBuildersToString(latestRestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestRestrictedConfig.ResourceName(), latestSlug, bucketSlug, "true"),
			},
			{ // Check drift mitigation for HCP managed channel from true->false
				PreConfig: func() {
					updateChannelRestriction(t, bucketSlug, latestSlug, false)
				},
				Config: testConfig(testAccConfigBuildersToString(latestRestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestRestrictedConfig.ResourceName(), latestSlug, bucketSlug, "true"),
			},
		},
	})
}

func testAccCheckPackerChannel(resourceName string, channelName string, bucketName string, restriction string) resource.TestCheckFunc {
	tests := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "author_id"),
		resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
		resource.TestCheckResourceAttrSet(resourceName, "created_at"),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", channelName),
		resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
		resource.TestCheckResourceAttrSet(resourceName, "managed"),
	}
	if restriction != "" {
		tests = append(tests, resource.TestCheckResourceAttr(resourceName, "restricted", restriction))
	} else {
		tests = append(tests, resource.TestCheckResourceAttrSet(resourceName, "restricted"))
	}
	return resource.ComposeAggregateTestCheckFunc(tests...)
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
