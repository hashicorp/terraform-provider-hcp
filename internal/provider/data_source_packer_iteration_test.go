// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_dataSourcePackerIteration_Simple(t *testing.T) {
	bucketSlug := testAccCreateSlug("IterationSimple")
	channelSlug := bucketSlug // No need for a different slug

	config := testAccPackerDataIterationBuilder("Simple", fmt.Sprintf("%q", bucketSlug), fmt.Sprintf("%q", channelSlug))

	var iteration *models.HashicorpCloudPackerIteration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			iteration, _ = upsertCompleteIteration(t, bucketSlug, "1234", nil)
			upsertChannel(t, bucketSlug, channelSlug, iteration.ID)
		},
		ProviderFactories: providerFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(config)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIterationState(config.BlockName(), bucketSlug, channelSlug),
					testAccCheckIterationStateMatchesIteration(config.BlockName(), &iteration),
				),
			},
			{
				// Test that references to `revoked_at` don't error when the
				// the iteration is not revoked or scheduled for revocation.
				//
				// Checked in this manner because you cant check for attributes
				// set to empty strings with `resource.TestCheckResourceAttr`
				Config: testConfig(testAccConfigBuildersToString(
					config,
					testAccOutputBuilder("revoke_at", config.AttributeRef("revoke_at")),
				)),
			},
		},
	})
}

func TestAcc_dataSourcePackerIteration_revokedIteration(t *testing.T) {
	bucketSlug := testAccCreateSlug("IterationRevoked")
	channelSlug := bucketSlug // No need for a different slug
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(24 * time.Hour))

	config := testAccPackerDataIterationBuilder("Revoked", fmt.Sprintf("%q", bucketSlug), fmt.Sprintf("%q", channelSlug))

	var revokedIteration *models.HashicorpCloudPackerIteration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			unrevokedIteration, _ := upsertCompleteIteration(t, bucketSlug, "1234", nil)
			upsertChannel(t, bucketSlug, channelSlug, unrevokedIteration.ID)
			revokedIteration = revokeIteration(t, unrevokedIteration.ID, bucketSlug, revokeAt)
		},
		ProviderFactories: providerFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(config)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIterationState(config.BlockName(), bucketSlug, channelSlug),
					resource.TestCheckResourceAttr(config.BlockName(), "revoke_at", revokeAt.String()),
					testAccCheckIterationStateMatchesIteration(config.BlockName(), &revokedIteration),
				),
			},
		},
	})
}

func testAccPackerDataIterationBuilder(uniqueName string, bucketName string, channelName string) testAccConfigBuilderInterface {
	return &testAccResourceConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_iteration",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"bucket_name": bucketName,
			"channel":     channelName,
		},
	}
}

func testAccCheckIterationState(resourceName, bucketName, channelName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
		resource.TestCheckResourceAttr(resourceName, "channel", channelName),
	)
}

func testAccCheckIterationStateMatchesIteration(resourceName string, iterationPtr **models.HashicorpCloudPackerIteration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		var iteration *models.HashicorpCloudPackerIteration
		if iterationPtr != nil {
			iteration = *iterationPtr
		}
		if iteration == nil {
			iteration = &models.HashicorpCloudPackerIteration{}
		}

		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, "bucket_name", iteration.BucketSlug),
			resource.TestCheckResourceAttr(resourceName, "ulid", iteration.ID),
			resource.TestCheckResourceAttr(resourceName, "fingerprint", iteration.Fingerprint),
			resource.TestCheckResourceAttr(resourceName, "incremental_version", fmt.Sprintf("%d", iteration.IncrementalVersion)),
			resource.TestCheckResourceAttr(resourceName, "author_id", iteration.AuthorID),
			resource.TestCheckResourceAttr(resourceName, "created_at", iteration.CreatedAt.String()),
			resource.TestCheckResourceAttr(resourceName, "updated_at", iteration.UpdatedAt.String()),
		}

		if !iteration.RevokeAt.IsZero() {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, "revoke_at", iteration.RevokeAt.String()))
		}

		return resource.ComposeAggregateTestCheckFunc(checks...)(state)
	}
}
