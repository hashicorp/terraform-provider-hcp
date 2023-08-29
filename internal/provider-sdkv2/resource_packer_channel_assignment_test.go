// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func TestAccPackerChannelAssignment_SimpleSetUnset(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentSimpleSetUnset")
	channelSlug := bucketSlug // No need for a different slug
	iterationFingerprint := "1"

	var iteration *models.HashicorpCloudPackerIteration

	baseAssignment := testAccPackerAssignmentBuilderBase("SimpleSetUnset", fmt.Sprintf("%q", bucketSlug), fmt.Sprintf("%q", channelSlug))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			upsertChannel(t, bucketSlug, channelSlug, "")
			iteration, _ = upsertCompleteIteration(t, bucketSlug, iterationFingerprint, nil)
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccCheckAssignmentDestroyed(baseAssignment.BlockName())(state); err != nil {
				t.Error(err)
			}
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Set channel assignment to the iteration
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignment(
					baseAssignment,
					``, fmt.Sprintf("%q", iterationFingerprint), ``,
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesIteration(baseAssignment.BlockName(), &iteration),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
			{ // Validate importing channel assignments that are already set
				ResourceName:      baseAssignment.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
			{ // Set channel assignment to null
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignment(
					baseAssignment,
					``, fmt.Sprintf("%q", unassignString), ``,
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesIteration(baseAssignment.BlockName(), nil),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
			{ // Validate importing channel assignments that are null
				ResourceName:      baseAssignment.BlockName(),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPackerChannelAssignment_AssignLatest(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentAssignLatest")
	channelSlug := bucketSlug // No need for a different slug
	uniqueName := "AssignLatest"

	// This config creates a data source that is read before apply time
	beforeIteration := testAccPackerDataIterationBuilder(
		uniqueName,
		fmt.Sprintf("%q", bucketSlug),
		`"latest"`,
	)
	beforeChannel := testAccPackerChannelBuilderBase(
		uniqueName,
		fmt.Sprintf("%q", channelSlug),
		beforeIteration.AttributeRef("bucket_name"),
	)
	beforeAssignment := testAccPackerAssignmentBuilderWithChannelReference(
		uniqueName,
		beforeChannel,
		beforeIteration.AttributeRef("id"), ``, ``,
	)

	// This config creates a data source that is read after apply time,
	// which is important for testing that CustomizeDiff doesn't cause errors
	afterChannel := testAccPackerChannelBuilderBase(
		uniqueName,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)
	afterIteration := testAccPackerDataIterationBuilder(
		uniqueName,
		afterChannel.AttributeRef("bucket_name"),
		`"latest"`,
	)
	afterAssignment := testAccPackerAssignmentBuilderWithChannelReference(
		uniqueName,
		afterChannel,
		afterIteration.AttributeRef("id"), ``, ``,
	)

	var iteration *models.HashicorpCloudPackerIteration

	generateStep := func(iterationData, channelResource, assignmentResource testAccConfigBuilderInterface) resource.TestStep {
		return resource.TestStep{
			Config: testConfig(testAccConfigBuildersToString(iterationData, channelResource, assignmentResource)),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckAssignmentStateBucketAndChannelName(assignmentResource.BlockName(), bucketSlug, channelSlug),
				testAccCheckAssignmentStateMatchesIteration(assignmentResource.BlockName(), &iteration),
				testAccCheckAssignmentStateMatchesChannelState(assignmentResource.BlockName(), channelResource.BlockName()),
				testAccCheckAssignmentStateMatchesAPI(assignmentResource.BlockName()),
			),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			iteration, _ = upsertCompleteIteration(t, bucketSlug, "abc", nil)
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			generateStep(beforeIteration, beforeChannel, beforeAssignment),
			{ // Remove any resources and data sources completely
				Config: testConfig(""),
			},
			generateStep(afterIteration, afterChannel, afterAssignment),
		},
	})
}

func TestAccPackerChannelAssignment_InvalidInputs(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentInvalidInputs")
	channelSlug := bucketSlug // No need for a different slug

	generateStep := func(iterID string, iterFingerprint string, iterVersion string, errorRegex string) resource.TestStep {
		return resource.TestStep{
			Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilder(
				"InvalidInputs",
				fmt.Sprintf("%q", bucketSlug),
				fmt.Sprintf("%q", channelSlug),
				iterID, iterFingerprint, iterVersion,
			))),
			ExpectError: regexp.MustCompile(errorRegex),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			upsertChannel(t, bucketSlug, channelSlug, "")
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			generateStep(
				`""`, ``, ``,
				`.*expected "iteration_id" to not be an empty string.*`,
			),
			generateStep(
				``, `""`, ``,
				`.*expected "iteration_fingerprint" to not be an empty string.*`,
			),
			generateStep(
				`"abcd"`, `"efgh"`, ``,
				`.*only one of.*\n.*can be specified.*`,
			),
			generateStep(
				`"1234"`, ``, `5678`,
				`.*only one of.*\n.*can be specified.*`,
			),
			generateStep(
				``, `"jamesBond"`, `007`,
				`.*only one of.*\n.*can be specified.*`,
			),
			generateStep(
				`"doesNotExist"`, ``, ``,
				`.*iteration with attributes \(id: doesNotExist\) does not exist.*`,
			),
			generateStep(
				``, `"alsoDoesNotExist"`, ``,
				`.*iteration with attributes \(fingerprint: alsoDoesNotExist\) does not exist.*`,
			),
			generateStep(
				``, ``, `99`,
				`.*iteration with attributes \(incremental_version: 99\) does not exist.*`,
			),
		},
	})
}

func TestAccPackerChannelAssignment_CreateFailsWhenPreassigned(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentCreateFailPreassign")
	channelSlug := bucketSlug // No need for a different slug
	iterationFingerprint := "1"

	channel := testAccPackerChannelBuilderBase(
		channelSlug,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)

	assignment := testAccPackerAssignmentBuilderWithChannelReference(
		"CreateFailsWhenPreassigned",
		channel,
		``, ``, `0`,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			upsertCompleteIteration(t, bucketSlug, iterationFingerprint, nil)
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConfigBuildersToString(channel)),
			},
			{
				PreConfig: func() {
					updateChannelAssignment(t,
						bucketSlug,
						channelSlug,
						&models.HashicorpCloudPackerIteration{Fingerprint: iterationFingerprint},
					)
				},
				Config:      testConfig(testAccConfigBuildersToString(channel, assignment)),
				ExpectError: regexp.MustCompile(".*channel with.*already has an assigned iteration.*"),
			},
		},
	})
}

func TestAccPackerChannelAssignment_HCPManagedChannelErrors(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentHCPManaged")
	channelSlug := "latest"

	assignment := testAccPackerAssignmentBuilder(
		"HCPManagedChannelErrors",
		fmt.Sprintf("%q", bucketSlug),
		fmt.Sprintf("%q", channelSlug),
		``, ``, `0`,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:      testConfig(testAccConfigBuildersToString(assignment)),
				ExpectError: regexp.MustCompile(".*channel with.*is managed by HCP Packer.*"),
			},
			{
				ResourceName:  assignment.BlockName(),
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s:%s", bucketSlug, channelSlug),
				ExpectError:   regexp.MustCompile(".*channel with.*is managed by HCP Packer.*"),
			},
			{ // Create a dummy non-empty state so that `CheckDestroy` will run.
				Config: testAccConfigDummyNonemptyState,
			},
		},
	})
}

// Test that all attributes generate and successfully apply plans to fix
// the assignment when it is changed OOB from null to a non-null iteration
func TestAccPackerChannelAssignment_EnforceNull(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentEnforceNull")
	channelSlug := bucketSlug // No need for a different slug

	channel := testAccPackerChannelBuilderBase(channelSlug,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)

	var iteration1 *models.HashicorpCloudPackerIteration
	var iteration2 *models.HashicorpCloudPackerIteration

	baseAssignment := testAccPackerAssignmentBuilderBaseWithChannelReference("EnforceNull", channel)

	generateEnforceNullCheckSteps := func(iterID string, iterFingerprint string, iterVersion string) []resource.TestStep {
		assignment := testAccPackerAssignmentBuilderFromAssignment(
			baseAssignment,
			iterID, iterFingerprint, iterVersion,
		)

		config := testConfig(testAccConfigBuildersToString(channel, assignment))

		checks := resource.ComposeAggregateTestCheckFunc(
			testAccCheckAssignmentStateBucketAndChannelName(assignment.BlockName(), bucketSlug, channelSlug),
			testAccCheckAssignmentStateMatchesIteration(assignment.BlockName(), nil),
			testAccCheckAssignmentStateMatchesChannelState(assignment.BlockName(), channel.BlockName()),
			testAccCheckAssignmentStateMatchesAPI(assignment.BlockName()),
		)

		return []resource.TestStep{
			{ // Set up channel and set the assignment using Terraform
				// This should be a no-op unless it is the first step, where the channel and null assignment are
				// initially created. However, this step is included every time to make sure we've applied the
				// assignment to the state at least once before checking if it is properly enforced against OOB changes.
				Config: config,
				Check:  checks,
			},
			{ // Change assignment OOB, then test with assignment set by Terraform
				PreConfig: func() {
					updateChannelAssignment(t, bucketSlug, channelSlug, &models.HashicorpCloudPackerIteration{ID: iteration1.ID})
					updateChannelAssignment(t, bucketSlug, channelSlug, &models.HashicorpCloudPackerIteration{ID: iteration2.ID})
				},
				Config: config,
				Check:  checks,
			},
		}
	}

	var generatedSteps []resource.TestStep
	// Add null ID steps
	generatedSteps = append(generatedSteps, generateEnforceNullCheckSteps(fmt.Sprintf("%q", unassignString), ``, ``)...)
	// Add null Fingerprint steps
	generatedSteps = append(generatedSteps, generateEnforceNullCheckSteps(``, fmt.Sprintf("%q", unassignString), ``)...)
	// Add null Version steps
	generatedSteps = append(generatedSteps, generateEnforceNullCheckSteps(``, ``, `0`)...)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			// Pushing two iterations so that we can also implicitly verify that
			// nullifying the assignment doesn't actually result in a rollback to iteration1
			iteration1, _ = upsertCompleteIteration(t, bucketSlug, "1", nil)
			iteration2, _ = upsertCompleteIteration(t, bucketSlug, "2", nil)
		},
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: generatedSteps,
	})
}

// An AssignmentBuilder without any iteration fields set.
// To be used downstream by other assignments to ensure core settings aren't changed.
func testAccPackerAssignmentBuilderBase(uniqueName string, bucketName string, channelName string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		uniqueName,
		bucketName,
		channelName,
		``, ``, ``,
	)
}

func testAccPackerAssignmentBuilderBaseWithChannelReference(uniqueName string, channel testAccConfigBuilderInterface) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilderBase(
		uniqueName,
		channel.AttributeRef("bucket_name"),
		channel.AttributeRef("name"),
	)
}

func testAccPackerAssignmentBuilder(uniqueName string, bucketName string, channelName string, iterID string, iterFingerprint string, iterVersion string) testAccConfigBuilderInterface {
	return &testAccResourceConfigBuilder{
		resourceType: "hcp_packer_channel_assignment",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"bucket_name":           bucketName,
			"channel_name":          channelName,
			"iteration_id":          iterID,
			"iteration_fingerprint": iterFingerprint,
			"iteration_version":     iterVersion,
		},
	}
}

func testAccPackerAssignmentBuilderFromAssignment(oldAssignment testAccConfigBuilderInterface, iterID string, iterFingerprint string, iterVersion string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		oldAssignment.UniqueName(),
		oldAssignment.Attributes()["bucket_name"],
		oldAssignment.Attributes()["channel_name"],
		iterID,
		iterFingerprint,
		iterVersion,
	)
}

func testAccPackerAssignmentBuilderWithChannelReference(uniqueName string, channel testAccConfigBuilderInterface, iterID string, iterFingerprint string, iterVersion string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		uniqueName,
		channel.AttributeRef("bucket_name"),
		channel.AttributeRef("name"),
		iterID, iterFingerprint, iterVersion,
	)
}

func testAccCheckAssignmentStateBucketAndChannelName(resourceName string, bucketName string, channelName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
		resource.TestCheckResourceAttr(resourceName, "channel_name", channelName),
	)
}

func testAccCheckAssignmentStateMatchesChannelState(assignmentResourceName string, channelResourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrPair(assignmentResourceName, "organization_id", channelResourceName, "organization_id"),
		resource.TestCheckResourceAttrPair(assignmentResourceName, "project_id", channelResourceName, "project_id"),
		resource.TestCheckResourceAttrPair(assignmentResourceName, "bucket_name", channelResourceName, "bucket_name"),
		resource.TestCheckResourceAttrPair(assignmentResourceName, "channel_name", channelResourceName, "name"),
	)
}

func testAccCheckAssignmentStateMatchesIteration(resourceName string, iterationPtr **models.HashicorpCloudPackerIteration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		var iteration *models.HashicorpCloudPackerIteration
		if iterationPtr != nil {
			iteration = *iterationPtr
		}

		if iteration == nil {
			iteration = &models.HashicorpCloudPackerIteration{}
		}

		iterID := iteration.ID
		if iterID == "" {
			iterID = unassignString
		}

		iterFingerprint := iteration.Fingerprint
		if iterFingerprint == "" {
			iterFingerprint = unassignString
		}

		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "iteration_id", iterID),
			resource.TestCheckResourceAttr(resourceName, "iteration_fingerprint", iterFingerprint),
			resource.TestCheckResourceAttr(resourceName, "iteration_version", fmt.Sprintf("%d", iteration.IncrementalVersion)),
		)(state)
	}
}

func testAccPullIterationFromAPIWithAssignmentState(resourceName string, state *terraform.State) (*models.HashicorpCloudPackerIteration, error) {
	client := testAccProvider.Meta().(*clients.Client)

	loc, _ := testAccGetLocationFromState(resourceName, state)

	bucketName, err := testAccGetAttributeFromResourceInState(resourceName, "bucket_name", state)
	if err != nil {
		return nil, err
	}
	channelName, err := testAccGetAttributeFromResourceInState(resourceName, "channel_name", state)
	if err != nil {
		return nil, err
	}

	channel, err := clients.GetPackerChannelBySlug(context.Background(), client, loc, *bucketName, *channelName)
	if err != nil {
		return nil, err
	}

	return channel.Iteration, nil
}

func testAccCheckAssignmentStateMatchesAPI(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		iteration, err := testAccPullIterationFromAPIWithAssignmentState(resourceName, state)
		if err != nil {
			return err
		}
		return testAccCheckAssignmentStateMatchesIteration(resourceName, &iteration)(state)
	}
}

func testAccCheckAssignmentDestroyed(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		iteration, err := testAccPullIterationFromAPIWithAssignmentState(resourceName, state)
		if err != nil {
			return fmt.Errorf("Unexpected error while validating channel assignment destruction. Got %v", err)
		} else if iteration != nil && (iteration.ID != "" || iteration.Fingerprint != "" || iteration.IncrementalVersion != 0) {
			return fmt.Errorf("Resource %q not properly destroyed", resourceName)
		}

		return nil
	}
}
