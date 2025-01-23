// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_Packer_Channel(t *testing.T) {
	t.Parallel()

	bucketName := testAccCreateSlug("ChannelSimple")
	channelName := bucketName // No need for a different name
	channelConfig := testAccPackerChannelBuilderBase("SimpleChannel", fmt.Sprintf("%q", channelName), fmt.Sprintf("%q", bucketName))
	unrestrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "false")
	restrictedChannelConfig := testAccPackerChannelBuilderFromChannel(channelConfig, "true")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketName)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketName, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(channelConfig)),
				Check:  testAccCheckPackerChannel(channelConfig.BlockName(), channelName, bucketName, ""),
			},
			{
				ResourceName:      channelConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
			{ // Unrestrict channel (likely a no-op)
				Config: testConfig(testAccConfigBuildersToString(unrestrictedChannelConfig)),
				Check:  testAccCheckPackerChannel(unrestrictedChannelConfig.BlockName(), channelName, bucketName, "false"),
			},
			{ // Validate importing explicitly unrestricted channel
				ResourceName:      unrestrictedChannelConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
			{ // Restrict channel
				Config: testConfig(testAccConfigBuildersToString(restrictedChannelConfig)),
				Check:  testAccCheckPackerChannel(restrictedChannelConfig.BlockName(), channelName, bucketName, "true"),
			},
			{ // Validate importing explicitly restricted channel
				ResourceName:      restrictedChannelConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_Packer_Channel_HCPManaged(t *testing.T) {
	t.Parallel()

	bucketName := testAccCreateSlug("ChannelHCPManaged")
	channelName := "latest"
	latestConfig := testAccPackerChannelBuilderBase("latest", fmt.Sprintf("%q", channelName), fmt.Sprintf("%q", bucketName))
	unrestrictedLatestConfig := testAccPackerChannelBuilderFromChannel(latestConfig, "false")
	restrictedLatestConfig := testAccPackerChannelBuilderFromChannel(latestConfig, "true")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketName)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketName, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Validate "creating" (automatically adopting) a managed channel
				Config: testConfig(testAccConfigBuildersToString(latestConfig)),
				Check:  testAccCheckPackerChannel(latestConfig.BlockName(), channelName, bucketName, ""),
			},
			{
				ResourceName:      latestConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
			{ // Unrestrict managed channel
				Config: testConfig(testAccConfigBuildersToString(unrestrictedLatestConfig)),
				Check:  testAccCheckPackerChannel(unrestrictedLatestConfig.BlockName(), channelName, bucketName, "false"),
			},
			{ // Validate importing explicitly unrestricted managed channel
				ResourceName:      unrestrictedLatestConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
			{ // Restrict managed channel
				Config: testConfig(testAccConfigBuildersToString(restrictedLatestConfig)),
				Check:  testAccCheckPackerChannel(restrictedLatestConfig.BlockName(), channelName, bucketName, "true"),
			},
			{ // Validate importing explicitly restricted managed channel
				ResourceName:      restrictedLatestConfig.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketName, channelName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_Packer_Channel_RestrictionDrift(t *testing.T) {
	t.Parallel()

	bucketName := testAccCreateSlug("RestrictionDrift")
	channelName := bucketName // No need for a different name

	channelUnrestrictedConfig := testAccPackerChannelBuilder("Drift", fmt.Sprintf("%q", channelName), fmt.Sprintf("%q", bucketName), "false")
	channelRestrictedConfig := testAccPackerChannelBuilderFromChannel(channelUnrestrictedConfig, "true")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketName)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketName, true)
			return nil
		},
		Steps: []resource.TestStep{
			// Normal channels
			{
				Config: testConfig(testAccConfigBuildersToString(channelUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelUnrestrictedConfig.BlockName(), channelName, bucketName, "false"),
			},
			{ // Check drift mitigation for normal channel from false->true
				PreConfig: func() {
					updateChannelRestriction(t, bucketName, channelName, true)
				},
				Config: testConfig(testAccConfigBuildersToString(channelUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelUnrestrictedConfig.BlockName(), channelName, bucketName, "false"),
			},
			{
				Config: testConfig(testAccConfigBuildersToString(channelRestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelRestrictedConfig.BlockName(), channelName, bucketName, "true"),
			},
			{ // Check drift mitigation for normal channel from true->false
				PreConfig: func() {
					updateChannelRestriction(t, bucketName, channelName, false)
				},
				Config: testConfig(testAccConfigBuildersToString(channelRestrictedConfig)),
				Check:  testAccCheckPackerChannel(channelRestrictedConfig.BlockName(), channelName, bucketName, "true"),
			},
		},
	})
}

func TestAcc_Packer_Channel_RestrictionDriftHCPManaged(t *testing.T) {
	t.Parallel()

	bucketName := testAccCreateSlug("RestrictionDriftHCPManaged")
	latestName := "latest"

	latestUnrestrictedConfig := testAccPackerChannelBuilder("DriftHCPManaged", fmt.Sprintf("%q", latestName), fmt.Sprintf("%q", bucketName), "false")
	latestRestrictedConfig := testAccPackerChannelBuilderFromChannel(latestUnrestrictedConfig, "true")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketName)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteBucket(t, bucketName, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(latestUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestUnrestrictedConfig.BlockName(), latestName, bucketName, "false"),
			},
			{ // Check drift mitigation for HCP managed channel from false->true
				PreConfig: func() {
					updateChannelRestriction(t, bucketName, latestName, true)
				},
				Config: testConfig(testAccConfigBuildersToString(latestUnrestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestUnrestrictedConfig.BlockName(), latestName, bucketName, "false"),
			},
			{
				Config: testConfig(testAccConfigBuildersToString(latestRestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestRestrictedConfig.BlockName(), latestName, bucketName, "true"),
			},
			{ // Check drift mitigation for HCP managed channel from true->false
				PreConfig: func() {
					updateChannelRestriction(t, bucketName, latestName, false)
				},
				Config: testConfig(testAccConfigBuildersToString(latestRestrictedConfig)),
				Check:  testAccCheckPackerChannel(latestRestrictedConfig.BlockName(), latestName, bucketName, "true"),
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
	return &testAccResourceConfigBuilder{
		resourceType: "hcp_packer_channel",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"name":        channelName,
			"bucket_name": bucketName,
			"restricted":  restricted,
		},
	}
}
