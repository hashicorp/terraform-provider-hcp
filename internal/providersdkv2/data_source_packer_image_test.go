// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_dataSourcePackerImage_Simple(t *testing.T) {
	bucketSlug := testAccCreateSlug("ImageSimple")
	channelSlug := bucketSlug // No need for a different slug

	buildOptions := buildOptions{
		cloudProvider: "aws",
		componentType: "amazon-ebs.example",
		images: []*models.HashicorpCloudPackerImageCreateBody{
			{
				ImageID: "image1",
				Region:  "us-east-1",
			},
			{
				ImageID: "image1",
				Region:  "us-west-1",
			},
		},
		labels: map[string]string{"test123": "test456"},
	}
	region := buildOptions.images[0].Region

	imageConfig := testAccPackerDataImageBuilder(
		"simple",
		fmt.Sprintf("%q", bucketSlug),
		fmt.Sprintf("%q", channelSlug),
		``,
		fmt.Sprintf("%q", buildOptions.cloudProvider),
		fmt.Sprintf("%q", region),
		fmt.Sprintf("%q", buildOptions.componentType),
	)
	errorImageConfig := testAccPackerDataImageBuilderFromImage(
		"error", imageConfig,
		fmt.Sprintf("%q", buildOptions.cloudProvider),
		fmt.Sprintf("%q", region),
		fmt.Sprintf("%q", "NotRealComponentType"),
	)

	var iteration *models.HashicorpCloudPackerIteration
	var build *models.HashicorpCloudPackerBuild

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			iteration, build = upsertCompleteIteration(t, bucketSlug, "1234", &buildOptions)
			upsertChannel(t, bucketSlug, channelSlug, iteration.ID)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(imageConfig)),
				Check:  testAccCheckPackerImageState(t, imageConfig.BlockName(), bucketSlug, &iteration, &build, region),
			},
			{
				// Test that references to `revoked_at` don't error when the
				// the iteration is not revoked or scheduled for revocation.
				//
				// Checked in this manner because you cant check for attributes
				// set to empty strings with `resource.TestCheckResourceAttr`
				Config: testConfig(testAccConfigBuildersToString(
					imageConfig,
					testAccOutputBuilder("revoke_at", imageConfig.AttributeRef("revoke_at")),
				)),
			},
			{ // Testing that filtering non-existent image fails properly
				PlanOnly:    true,
				Config:      testConfig(testAccConfigBuildersToString(errorImageConfig)),
				ExpectError: regexp.MustCompile("Could not find image"),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_IterationID(t *testing.T) {
	bucketSlug := testAccCreateSlug("ImageFromIterationId")
	channelSlug := bucketSlug // No need for a different slug

	buildOptions := defaultBuildOptions
	region := buildOptions.images[0].Region

	iterationConfig := testAccPackerDataIterationBuilder(
		"iteration",
		fmt.Sprintf("%q", bucketSlug),
		fmt.Sprintf("%q", channelSlug),
	)
	imageConfig := testAccPackerDataImageBuilderWithIterationReference(
		"from_iteration_id",
		iterationConfig,
		fmt.Sprintf("%q", buildOptions.cloudProvider),
		fmt.Sprintf("%q", region),
		fmt.Sprintf("%q", buildOptions.componentType),
	)

	var iteration *models.HashicorpCloudPackerIteration
	var build *models.HashicorpCloudPackerBuild

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			iteration, build = upsertCompleteIteration(t, bucketSlug, "1234", &buildOptions)
			upsertChannel(t, bucketSlug, channelSlug, iteration.ID)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				// Check that everything works when using `iteration_id` since
				// the simple test uses `channel` to avoid having a data source
				Config: testConfig(testAccConfigBuildersToString(iterationConfig, imageConfig)),
				Check:  testAccCheckPackerImageState(t, imageConfig.BlockName(), bucketSlug, &iteration, &build, region),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_revokedIteration(t *testing.T) {
	bucketSlug := testAccCreateSlug("ImageRevoked")
	channelSlug := bucketSlug // No need for a different slug

	revokeAt := strfmt.DateTime(time.Now().UTC().Add(24 * time.Hour))

	buildOptions := defaultBuildOptions
	region := buildOptions.images[0].Region
	config := testAccPackerDataImageBuilder(
		"revoked",
		fmt.Sprintf("%q", bucketSlug),
		fmt.Sprintf("%q", channelSlug),
		``,
		fmt.Sprintf("%q", buildOptions.cloudProvider),
		fmt.Sprintf("%q", region),
		fmt.Sprintf("%q", buildOptions.componentType),
	)

	var iteration *models.HashicorpCloudPackerIteration
	var build *models.HashicorpCloudPackerBuild

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			iteration, build = upsertCompleteIteration(t, bucketSlug, "1234", &buildOptions)
			upsertChannel(t, bucketSlug, channelSlug, iteration.ID)
			revokeIteration(t, iteration.ID, bucketSlug, revokeAt)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(config)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackerImageState(t, config.BlockName(), bucketSlug, &iteration, &build, region),
					resource.TestCheckResourceAttr(config.BlockName(), "revoke_at", revokeAt.String()),
				),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_emptyChannel(t *testing.T) {
	bucketSlug := testAccCreateSlug("ImageEmptyChannel")
	channelSlug := bucketSlug // No need for a different slug

	config := testAccPackerDataImageBuilder(
		"empty_channel",
		fmt.Sprintf("%q", bucketSlug),
		fmt.Sprintf("%q", channelSlug),
		``,
		fmt.Sprintf("%q", "someProvider"),
		fmt.Sprintf("%q", "someRegion"),
		``,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			upsertChannel(t, bucketSlug, channelSlug, "")
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:      testConfig(testAccConfigBuildersToString(config)),
				ExpectError: regexp.MustCompile(`.*Channel does not have an assigned iteration.*`),
			},
			{ // Create a dummy non-empty state so that `CheckDestroy` will run.
				Config: testAccConfigDummyNonemptyState,
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_channelAndIterationIDReject(t *testing.T) {
	config := testAccPackerDataImageBuilder(
		"invalid_args",
		fmt.Sprintf("%q", "someBucketSlug"),
		fmt.Sprintf("%q", "someChannelSlug"),
		fmt.Sprintf("%q", "someIterationID"),
		fmt.Sprintf("%q", "someProvider"),
		fmt.Sprintf("%q", "someRegion"),
		``,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testConfig(testAccConfigBuildersToString(config)),
				ExpectError: regexp.MustCompile("Error: Invalid combination of arguments"),
			},
		},
	})
}

func testAccPackerDataImageBuilder(uniqueName, bucketName,
	channelName, iterationID,
	cloudProvider, region, componentType string) testAccConfigBuilderInterface {

	return &testAccResourceConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_image",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"bucket_name":    bucketName,
			"channel":        channelName,
			"iteration_id":   iterationID,
			"cloud_provider": cloudProvider,
			"region":         region,
			"component_type": componentType,
		},
	}
}

func testAccPackerDataImageBuilderWithIterationReference(uniqueName string, iteration testAccConfigBuilderInterface,
	cloudProvider, region, componentType string) testAccConfigBuilderInterface {

	return testAccPackerDataImageBuilder(
		uniqueName,
		iteration.AttributeRef("bucket_name"),
		``,
		iteration.AttributeRef("ulid"),
		cloudProvider,
		region,
		componentType,
	)
}

func testAccPackerDataImageBuilderFromImage(uniqueName string, image testAccConfigBuilderInterface,
	cloudProvider, region, componentType string) testAccConfigBuilderInterface {

	return testAccPackerDataImageBuilder(
		uniqueName,
		image.Attributes()["bucket_name"],
		image.Attributes()["channel"],
		image.Attributes()["iteration_id"],
		cloudProvider,
		region,
		componentType,
	)
}

func testAccCheckPackerImageState(t *testing.T, resourceName, bucketName string, iterationPtr **models.HashicorpCloudPackerIteration,
	buildPtr **models.HashicorpCloudPackerBuild, region string) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		var iteration *models.HashicorpCloudPackerIteration
		if iterationPtr != nil {
			iteration = *iterationPtr
		}
		if iteration == nil {
			iteration = &models.HashicorpCloudPackerIteration{}
		}

		var build *models.HashicorpCloudPackerBuild
		if buildPtr != nil {
			build = *buildPtr
		}
		if build == nil {
			build = &models.HashicorpCloudPackerBuild{}
		}

		var matchingImages []*models.HashicorpCloudPackerImage
		for _, image := range build.Images {
			if image.Region == region {
				matchingImages = append(matchingImages, image)
			}
		}
		if len(matchingImages) == 0 {
			return fmt.Errorf("didn't find any images in the provided build in the specified region")
		}
		var image = matchingImages[0]
		if len(matchingImages) > 1 {
			t.Logf("found %d images in the provided build in the specified region, the first will be used", len(matchingImages))
		}

		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
			resource.TestCheckResourceAttrSet(resourceName, "project_id"),
			resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
			resource.TestCheckResourceAttr(resourceName, "iteration_id", iteration.ID),
			resource.TestCheckResourceAttr(resourceName, "build_id", build.ID),
			resource.TestCheckResourceAttr(resourceName, "id", image.ID),
			resource.TestCheckResourceAttr(resourceName, "packer_run_uuid", build.PackerRunUUID),
			resource.TestCheckResourceAttr(resourceName, "cloud_provider", build.CloudProvider),
			resource.TestCheckResourceAttr(resourceName, "component_type", build.ComponentType),
			resource.TestCheckResourceAttr(resourceName, "region", image.Region),
			resource.TestCheckResourceAttr(resourceName, "cloud_image_id", image.ImageID),
			resource.TestCheckResourceAttr(resourceName, "created_at", image.CreatedAt.String()),
		}

		if !iteration.RevokeAt.IsZero() {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, "revoke_at", iteration.RevokeAt.String()))
		}

		for key, value := range build.Labels {
			checks = append(checks, resource.TestCheckResourceAttr(
				resourceName, fmt.Sprintf("labels.%s", key), value,
			))
		}

		return resource.ComposeAggregateTestCheckFunc(checks...)(state)
	}
}
