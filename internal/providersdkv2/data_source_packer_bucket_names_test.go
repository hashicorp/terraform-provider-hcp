// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_Packer_dataSourcePackerBucketNames(t *testing.T) {
	bucket0 := testAccCreateSlug("1-BucketNames")
	bucket1 := testAccCreateSlug("2-BucketNames")
	bucket2 := testAccCreateSlug("3-BucketNames")

	bucketNames := testAccPackerDataBucketNamesBuilder("all")
	config := testConfig(testAccConfigBuildersToString(bucketNames))
	dataAddr := bucketNames.BlockName()

	// Assertions use only our slug names, not names.#, so this test stays valid when
	// other TestAcc_Packer_* run in parallel (-parallel=N) against the same HCP project.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataAddr, "id"),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket0),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket1),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket2),
				),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket0)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(dataAddr, "names.*", bucket0),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket1),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket2),
				),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket1)
					upsertBucket(t, bucket2)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(dataAddr, "names.*", bucket0),
					resource.TestCheckTypeSetElemAttr(dataAddr, "names.*", bucket1),
					resource.TestCheckTypeSetElemAttr(dataAddr, "names.*", bucket2),
				),
			},
			{
				PreConfig: func() {
					deleteBucket(t, bucket0, true)
					deleteBucket(t, bucket1, true)
					deleteBucket(t, bucket2, true)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket0),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket1),
					testAccCheckPackerBucketNamesNotContains(dataAddr, bucket2),
				),
			},
		},
	})
}

func testAccCheckPackerBucketNamesNotContains(addr, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource not found: %s", addr)
		}
		if rs.Primary == nil {
			return fmt.Errorf("no primary instance: %s", addr)
		}
		for k, v := range rs.Primary.Attributes {
			if !strings.HasPrefix(k, "names.") || k == "names.#" {
				continue
			}
			if v == bucketName {
				return fmt.Errorf("%s: expected names set not to include %q, but it does (attribute %q)", addr, bucketName, k)
			}
		}
		return nil
	}
}

func testAccPackerDataBucketNamesBuilder(uniqueName string) testAccConfigBuilderInterface {
	return testAccResourceConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_bucket_names",
		uniqueName:   uniqueName,
	}
}
