// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func TestAccPackerChannelAssignment_SimpleSetUnset(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentSimpleSetUnset")
	channelSlug := bucketSlug // No need for a different slug
	versionFingerprint := "1"

	var version *models.HashicorpCloudPacker20230101Version

	baseAssignment := testAccPackerAssignmentBuilderBase("SimpleSetUnset", fmt.Sprintf("%q", bucketSlug), fmt.Sprintf("%q", channelSlug))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			createOrUpdateChannel(t, bucketSlug, channelSlug, "")
			version, _ = upsertCompleteVersion(t, bucketSlug, versionFingerprint, nil)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccCheckAssignmentDestroyed(baseAssignment.BlockName())(state); err != nil {
				t.Error(err)
			}
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Set channel assignment to the version
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignment(
					baseAssignment,
					fmt.Sprintf("%q", versionFingerprint),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), &version),
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
					fmt.Sprintf("%q", unassignString),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), nil),
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
	beforeVersion := testAccPackerDataIterationBuilder(
		uniqueName,
		fmt.Sprintf("%q", bucketSlug),
		`"latest"`,
	)
	beforeChannel := testAccPackerChannelBuilderBase(
		uniqueName,
		fmt.Sprintf("%q", channelSlug),
		beforeVersion.AttributeRef("bucket_name"),
	)
	beforeAssignment := testAccPackerAssignmentBuilderWithChannelReference(
		uniqueName,
		beforeChannel,
		beforeVersion.AttributeRef("fingerprint"),
	)

	// This config creates a data source that is read after apply time,
	// which is important for testing that CustomizeDiff doesn't cause errors
	afterChannel := testAccPackerChannelBuilderBase(
		uniqueName,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)
	afterVersion := testAccPackerDataIterationBuilder(
		uniqueName,
		afterChannel.AttributeRef("bucket_name"),
		`"latest"`,
	)
	afterAssignment := testAccPackerAssignmentBuilderWithChannelReference(
		uniqueName,
		afterChannel,
		afterVersion.AttributeRef("fingerprint"),
	)

	var version *models.HashicorpCloudPacker20230101Version

	generateStep := func(versionData, channelResource, assignmentResource testAccConfigBuilderInterface) resource.TestStep {
		return resource.TestStep{
			Config: testConfig(testAccConfigBuildersToString(versionData, channelResource, assignmentResource)),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckAssignmentStateBucketAndChannelName(assignmentResource.BlockName(), bucketSlug, channelSlug),
				testAccCheckAssignmentStateMatchesVersion(assignmentResource.BlockName(), &version),
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
			version, _ = upsertCompleteVersion(t, bucketSlug, "abc", nil)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			generateStep(beforeVersion, beforeChannel, beforeAssignment),
			{ // Remove any resources and data sources completely
				Config: testConfig(""),
			},
			generateStep(afterVersion, afterChannel, afterAssignment),
		},
	})
}

func TestAccPackerChannelAssignment_InvalidInputs(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentInvalidInputs")
	channelSlug := bucketSlug // No need for a different slug

	generateStep := func(fingerprint string, errorRegex string) resource.TestStep {
		return resource.TestStep{
			Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilder(
				"InvalidInputs",
				fmt.Sprintf("%q", bucketSlug),
				fmt.Sprintf("%q", channelSlug),
				fingerprint,
			))),
			ExpectError: regexp.MustCompile(errorRegex),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			createOrUpdateChannel(t, bucketSlug, channelSlug, "")
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			generateStep(
				`""`,
				`.*expected "version_fingerprint" to not be an empty string.*`,
			),
			generateStep(
				`"doesNotExist"`,
				// TODO: Update to "version" once the new API is implemented for this resource.
				`.*version with fingerprint \(fingerprint: doesNotExist\) does not exist.*`,
			),
			// TODO: Remove once the iteration_fingerprint attribute is removed
			{
				Config: fmt.Sprintf(`
resource "hcp_packer_channel_assignment" "InvalidInputs" {
	bucket_name = %q
	channel_name = %q
	iteration_fingerprint = "someVersion"
	version_fingerprint = "someVersion"
}
			`, bucketSlug, channelSlug),
				ExpectError: regexp.MustCompile(`.*only one of.*\n.*can be specified.*`),
			},
			{ // Create a dummy non-empty state so that `CheckDestroy` will run.
				Config: testAccConfigDummyNonemptyState,
			},
		},
	})
}

func TestAccPackerChannelAssignment_CreateFailsWhenPreassigned(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentCreateFailPreassign")
	channelSlug := bucketSlug // No need for a different slug
	versionFingerprint := "1"

	channel := testAccPackerChannelBuilderBase(
		channelSlug,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)

	assignment := testAccPackerAssignmentBuilderWithChannelReference(
		"CreateFailsWhenPreassigned",
		channel,
		`"2"`,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			upsertCompleteVersion(t, bucketSlug, versionFingerprint, nil)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
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
						versionFingerprint,
					)
				},
				Config:      testConfig(testAccConfigBuildersToString(channel, assignment)),
				ExpectError: regexp.MustCompile(".*channel with.*already has an assigned version.*"),
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
		`1`,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
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
// the assignment when it is changed OOB from null to a non-null version
func TestAccPackerChannelAssignment_EnforceNull(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentEnforceNull")
	channelSlug := bucketSlug // No need for a different slug

	channel := testAccPackerChannelBuilderBase(channelSlug,
		fmt.Sprintf("%q", channelSlug),
		fmt.Sprintf("%q", bucketSlug),
	)

	var version1 *models.HashicorpCloudPacker20230101Version
	var version2 *models.HashicorpCloudPacker20230101Version

	baseAssignment := testAccPackerAssignmentBuilderBaseWithChannelReference("EnforceNull", channel)

	assignment := testAccPackerAssignmentBuilderFromAssignment(
		baseAssignment,
		fmt.Sprintf("%q", unassignString),
	)

	config := testConfig(testAccConfigBuildersToString(channel, assignment))

	checks := resource.ComposeAggregateTestCheckFunc(
		testAccCheckAssignmentStateBucketAndChannelName(assignment.BlockName(), bucketSlug, channelSlug),
		testAccCheckAssignmentStateMatchesVersion(assignment.BlockName(), nil),
		testAccCheckAssignmentStateMatchesChannelState(assignment.BlockName(), channel.BlockName()),
		testAccCheckAssignmentStateMatchesAPI(assignment.BlockName()),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			// Pushing two versions so that we can also implicitly verify that
			// nullifying the assignment doesn't actually result in a rollback to version1
			version1, _ = upsertCompleteVersion(t, bucketSlug, "1", nil)
			version2, _ = upsertCompleteVersion(t, bucketSlug, "2", nil)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Set up channel and set the assignment using Terraform
				// This should be a no-op unless it is the first step, where the channel and null assignment are
				// initially created. However, this step is included every time to make sure we've applied the
				// assignment to the state at least once before checking if it is properly enforced against OOB changes.
				Config: config,
				Check:  checks,
			},
			{ // Change assignment OOB, then test with assignment set by Terraform
				PreConfig: func() {
					updateChannelAssignment(t, bucketSlug, channelSlug, version1.Fingerprint)
					updateChannelAssignment(t, bucketSlug, channelSlug, version2.Fingerprint)
				},
				Config: config,
				Check:  checks,
			},
		},
	})
}

// TODO: Remove once the iteration_fingerprint attribute is removed
// Test that migration from `iteration_fingerprint` to `version_fingerprint` works properly
func TestAccPackerChannelAssignment_AliasMigration(t *testing.T) {
	bucketSlug := testAccCreateSlug("AssignmentAliasMigration")
	channelSlug := bucketSlug // No need for a different slug
	versionFingerprint := "1"

	var version *models.HashicorpCloudPacker20230101Version

	baseAssignment := testAccPackerAssignmentBuilderBase("AliasMigration", fmt.Sprintf("%q", bucketSlug), fmt.Sprintf("%q", channelSlug))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
			upsertBucket(t, bucketSlug)
			createOrUpdateChannel(t, bucketSlug, channelSlug, "")
			version, _ = upsertCompleteVersion(t, bucketSlug, versionFingerprint, nil)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			deleteBucket(t, bucketSlug, true)
			return nil
		},
		Steps: []resource.TestStep{
			{ // Set channel assignment to the version using `iteration_fingerprint`
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignmentWithIterationFingerprint(
					baseAssignment,
					fmt.Sprintf("%q", versionFingerprint),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), &version),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
			{ // Set channel assignment to the version using `version_fingerprint`
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignment(
					baseAssignment,
					fmt.Sprintf("%q", versionFingerprint),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), &version),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
			{ // Set channel assignment to null with `iteration_fingerprint`
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignmentWithIterationFingerprint(
					baseAssignment,
					fmt.Sprintf("%q", unassignString),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), nil),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
			{ // Set channel assignment to null with `version_fingerprint`
				Config: testConfig(testAccConfigBuildersToString(testAccPackerAssignmentBuilderFromAssignment(
					baseAssignment,
					fmt.Sprintf("%q", unassignString),
				))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssignmentStateBucketAndChannelName(baseAssignment.BlockName(), bucketSlug, channelSlug),
					testAccCheckAssignmentStateMatchesVersion(baseAssignment.BlockName(), nil),
					testAccCheckAssignmentStateMatchesAPI(baseAssignment.BlockName()),
				),
			},
		},
	})
}

// An AssignmentBuilder without any version fields set.
// To be used downstream by other assignments to ensure core settings aren't changed.
func testAccPackerAssignmentBuilderBase(uniqueName string, bucketName string, channelName string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		uniqueName,
		bucketName,
		channelName,
		``,
	)
}

func testAccPackerAssignmentBuilderBaseWithChannelReference(uniqueName string, channel testAccConfigBuilderInterface) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilderBase(
		uniqueName,
		channel.AttributeRef("bucket_name"),
		channel.AttributeRef("name"),
	)
}

func testAccPackerAssignmentBuilder(uniqueName string, bucketName string, channelName string, fingerprint string) testAccConfigBuilderInterface {
	return &testAccResourceConfigBuilder{
		resourceType: "hcp_packer_channel_assignment",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"bucket_name":         bucketName,
			"channel_name":        channelName,
			"version_fingerprint": fingerprint,
		},
	}
}

func testAccPackerAssignmentBuilderFromAssignment(oldAssignment testAccConfigBuilderInterface, fingerprint string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		oldAssignment.UniqueName(),
		oldAssignment.Attributes()["bucket_name"],
		oldAssignment.Attributes()["channel_name"],
		fingerprint,
	)
}

// TODO: Remove once the iteration_fingerprint attribute is removed
func testAccPackerAssignmentBuilderFromAssignmentWithIterationFingerprint(oldAssignment testAccConfigBuilderInterface, fingerprint string) testAccConfigBuilderInterface {
	return &testAccResourceConfigBuilder{
		resourceType: "hcp_packer_channel_assignment",
		uniqueName:   oldAssignment.UniqueName(),
		attributes: map[string]string{
			"bucket_name":           oldAssignment.Attributes()["bucket_name"],
			"channel_name":          oldAssignment.Attributes()["channel_name"],
			"iteration_fingerprint": fingerprint,
		},
	}
}

func testAccPackerAssignmentBuilderWithChannelReference(uniqueName string, channel testAccConfigBuilderInterface, fingerprint string) testAccConfigBuilderInterface {
	return testAccPackerAssignmentBuilder(
		uniqueName,
		channel.AttributeRef("bucket_name"),
		channel.AttributeRef("name"),
		fingerprint,
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

func testAccCheckAssignmentStateMatchesVersion(resourceName string, versionPtr **models.HashicorpCloudPacker20230101Version) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		var version *models.HashicorpCloudPacker20230101Version
		if versionPtr != nil {
			version = *versionPtr
		}

		if version == nil {
			version = &models.HashicorpCloudPacker20230101Version{}
		}

		versionFingerprint := version.Fingerprint
		if versionFingerprint == "" {
			versionFingerprint = unassignString
		}

		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "version_fingerprint", versionFingerprint),
			resource.TestCheckResourceAttr(resourceName, "iteration_fingerprint", versionFingerprint),
		)(state)
	}
}

func testAccPullVersionFromAPIWithAssignmentState(resourceName string, state *terraform.State) (*models.HashicorpCloudPacker20230101Version, error) {
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

	bucketLocation := location.GenericBucketLocation{
		Location: location.GenericLocation{
			OrganizationID: loc.OrganizationID,
			ProjectID:      loc.ProjectID,
		},
		BucketName: *bucketName,
	}
	channel, err := packerv2.GetChannelByName(client, bucketLocation, *channelName)
	if err != nil {
		return nil, err
	}

	return channel.Version, nil
}

func testAccCheckAssignmentStateMatchesAPI(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		version, err := testAccPullVersionFromAPIWithAssignmentState(resourceName, state)
		if err != nil {
			return err
		}
		return testAccCheckAssignmentStateMatchesVersion(resourceName, &version)(state)
	}
}

func testAccCheckAssignmentDestroyed(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		version, err := testAccPullVersionFromAPIWithAssignmentState(resourceName, state)
		if err != nil {
			return fmt.Errorf("Unexpected error while validating channel assignment destruction. Got %v", err)
		} else if version != nil && (version.ID != "" || version.Fingerprint != "" || getVersionNumber(version.Name) != 0) {
			return fmt.Errorf("Resource %q not properly destroyed", resourceName)
		}

		return nil
	}
}
