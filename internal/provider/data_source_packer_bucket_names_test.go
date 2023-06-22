package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_dataSourcePackerBucketNames(t *testing.T) {
	bucket0 := testAccCreateSlug("1-BucketNames")
	bucket1 := testAccCreateSlug("2-BucketNames")
	bucket2 := testAccCreateSlug("3-BucketNames")

	bucketNames := testAccPackerBucketNamesBuilder("all")
	config := testConfig(testAccConfigBuildersToString(bucketNames))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				// If this check fails, there are probably pre-existing buckets
				// in the environment that need to be deleted before testing.
				// This may also be caused by other tests failing to clean up properly.
				Check: resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.#", "0"),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket0)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.#", "1"),
					resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.0", bucket0),
				),
			},
			{
				PreConfig: func() {
					upsertBucket(t, bucket1)
					upsertBucket(t, bucket2)
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.#", "3"),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.0", bucket0),
						resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.1", bucket1),
						resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.2", bucket2),
					),
				),
			},
			{
				PreConfig: func() {
					deleteBucket(t, bucket0, true)
					deleteBucket(t, bucket1, true)
					deleteBucket(t, bucket2, true)
				},
				Config: config,
				Check:  resource.TestCheckResourceAttr(bucketNames.ResourceName(), "names.#", "0"),
			},
		},
	})
}

func testAccPackerBucketNamesBuilder(uniqueName string) testAccConfigBuilderInterface {
	return testAccConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_bucket_names",
		uniqueName:   uniqueName,
	}
}
