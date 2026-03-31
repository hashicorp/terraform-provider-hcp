// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"strconv"
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

	// Must not be Parallel: steps share API state and compare list length deltas.
	var baselineCount int

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
					testAccCheckPackerBucketNamesCountCapture(dataAddr, &baselineCount),
				),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket0)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackerBucketNamesCountEquals(dataAddr, baselineCount+1),
					resource.TestCheckTypeSetElemAttr(dataAddr, "names.*", bucket0),
				),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket1)
					upsertBucket(t, bucket2)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackerBucketNamesCountEquals(dataAddr, baselineCount+3),
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
					testAccCheckPackerBucketNamesCountEquals(dataAddr, baselineCount),
				),
			},
		},
	})
}

func testAccCheckPackerBucketNamesCountCapture(addr string, dest *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		n, err := testAccPackerBucketNamesCount(s, addr)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	}
}

func testAccCheckPackerBucketNamesCountEquals(addr string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := testAccPackerBucketNamesCount(s, addr)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("%s: names.# expected %d, got %d", addr, want, got)
		}
		return nil
	}
}

func testAccPackerBucketNamesCount(s *terraform.State, addr string) (int, error) {
	rs, ok := s.RootModule().Resources[addr]
	if !ok {
		return 0, fmt.Errorf("resource not found: %s", addr)
	}
	if rs.Primary == nil {
		return 0, fmt.Errorf("no primary instance: %s", addr)
	}
	raw, ok := rs.Primary.Attributes["names.#"]
	if !ok {
		return 0, fmt.Errorf("%s: names.# not in state", addr)
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse names.#: %w", err)
	}
	return n, nil
}

func testAccPackerDataBucketNamesBuilder(uniqueName string) testAccConfigBuilderInterface {
	return testAccResourceConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_bucket_names",
		uniqueName:   uniqueName,
	}
}
