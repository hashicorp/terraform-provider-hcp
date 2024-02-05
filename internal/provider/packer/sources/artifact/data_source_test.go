// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package artifact_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder/packerconfig"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testcheck"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testclient"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func TestAcc_Packer_Data_Artifact(t *testing.T) {
	// This is also checked further inside resource.ParallelTest, but we need to
	// check it here because the next like DefaultProjectLocation tries to create the provider
	// client, which it doesn't work in all evirnoments.
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}

	loc := acctest.DefaultProjectLocation(t)

	bucketName := testutils.CreateTestSlug("ArtifactSimple")
	channelName := bucketName // No need for a different slug
	emptyChannel := "empty"

	bucketLoc := location.GenericBucketLocation{
		Location:   loc,
		BucketName: bucketName,
	}

	fingerprint := acctest.RandString(32)

	buildOptions := testclient.UpsertBuildOptions{
		Complete:      true,
		Platform:      "aws",
		ComponentType: "amazon-ebs.example",
		Artifacts: []*packerv2.CreateArtifactBody{
			{
				Region: "us-east-1",
			},
			{
				Region: "us-west-1",
			},
		},
		Labels: map[string]string{"test123": "test456"},
	}
	region := buildOptions.Artifacts[0].Region

	artifactConfig := packerconfig.NewArtifactDataSourceBuilder("simple")
	artifactConfig.SetBucketName(fmt.Sprintf("%q", bucketName))
	artifactConfig.SetChannelName(fmt.Sprintf("%q", channelName))
	artifactConfig.SetPlatform(fmt.Sprintf("%q", buildOptions.Platform))
	artifactConfig.SetRegion(fmt.Sprintf("%q", region))
	artifactConfig.SetComponentType(fmt.Sprintf("%q", buildOptions.ComponentType))

	artifactConfigEmptyChannel := packerconfig.NewArtifactDataSourceBuilder("simple")
	artifactConfigEmptyChannel.SetBucketName(fmt.Sprintf("%q", bucketName))
	artifactConfigEmptyChannel.SetChannelName(fmt.Sprintf("%q", emptyChannel))
	artifactConfigEmptyChannel.SetPlatform(fmt.Sprintf("%q", buildOptions.Platform))
	artifactConfigEmptyChannel.SetRegion(fmt.Sprintf("%q", region))
	artifactConfigEmptyChannel.SetComponentType(fmt.Sprintf("%q", buildOptions.ComponentType))

	latestChannel := "latest"
	artifactConfigLatestChannel := packerconfig.NewArtifactDataSourceBuilder("simple")
	artifactConfigLatestChannel.SetBucketName(fmt.Sprintf("%q", bucketName))
	artifactConfigLatestChannel.SetChannelName(fmt.Sprintf("%q", latestChannel))
	artifactConfigLatestChannel.SetPlatform(fmt.Sprintf("%q", buildOptions.Platform))
	artifactConfigLatestChannel.SetRegion(fmt.Sprintf("%q", region))
	artifactConfigLatestChannel.SetComponentType(fmt.Sprintf("%q", buildOptions.ComponentType))

	errorArtifactConfig := packerconfig.NewArtifactDataSourceBuilder("error")
	errorArtifactConfig.SetBucketName(fmt.Sprintf("%q", bucketName))
	errorArtifactConfig.SetChannelName(fmt.Sprintf("%q", channelName))
	errorArtifactConfig.SetPlatform(fmt.Sprintf("%q", buildOptions.Platform))
	errorArtifactConfig.SetRegion(fmt.Sprintf("%q", region))
	errorArtifactConfig.SetComponentType(fmt.Sprintf("%q", "NotRealComponentType"))

	artifactConfigWithFingerprint := packerconfig.NewArtifactDataSourceBuilder("simple")
	artifactConfigWithFingerprint.SetBucketName(fmt.Sprintf("%q", bucketName))
	artifactConfigWithFingerprint.SetVersionFingerprint(fmt.Sprintf("%q", fingerprint))
	artifactConfigWithFingerprint.SetPlatform(fmt.Sprintf("%q", buildOptions.Platform))
	artifactConfigWithFingerprint.SetRegion(fmt.Sprintf("%q", region))
	artifactConfigWithFingerprint.SetComponentType(fmt.Sprintf("%q", buildOptions.ComponentType))

	scheduledRevokeFingerprint := acctest.RandString(32)
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(24 * time.Hour))

	var scheduledRevokeVersion *packerv2.Version
	var scheduledRevokeBuild *packerv2.Build
	var version *packerv2.Version
	var build *packerv2.Build

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testclient.UpsertRegistry(t, loc, nil)
			testclient.UpsertBucket(t, loc, bucketName)
			version, build = testclient.UpsertCompleteVersion(t, bucketLoc, fingerprint, &buildOptions)
			testclient.UpsertChannel(t, bucketLoc, channelName, fingerprint)
			testclient.UpsertChannel(t, bucketLoc, emptyChannel, "")
			_, scheduledRevokeBuild = testclient.UpsertCompleteVersion(t, bucketLoc, scheduledRevokeFingerprint, &buildOptions)
			scheduledRevokeVersion = testclient.ScheduleRevokeVersion(t, bucketLoc, scheduledRevokeFingerprint, revokeAt)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			if err := testclient.DeleteBucket(t, loc, bucketName); err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: configbuilder.BuildersToString(artifactConfig),
				Check:  checkDataSource(t, artifactConfig, bucketLoc, &version, &build, region),
			},
			{
				Config: configbuilder.BuildersToString(artifactConfigWithFingerprint),
				Check:  checkDataSource(t, artifactConfig, bucketLoc, &version, &build, region),
			},
			{
				Config: configbuilder.BuildersToString(artifactConfigLatestChannel),
				Check: resource.ComposeTestCheckFunc(
					checkDataSource(t, artifactConfigLatestChannel, bucketLoc, &scheduledRevokeVersion, &scheduledRevokeBuild, region),
					testcheck.Attribute(artifactConfigLatestChannel, "revoke_at", revokeAt.String()),
				),
			},
			{ // Testing that filtering non-existent artifact fails properly
				PlanOnly:    true,
				Config:      configbuilder.BuildersToString(errorArtifactConfig),
				ExpectError: regexp.MustCompile("Could not find an Artifact"),
			},
			{ // Testing that filtering empty channel fails properly
				PlanOnly:    true,
				Config:      configbuilder.BuildersToString(artifactConfigEmptyChannel),
				ExpectError: regexp.MustCompile("provided Channel does not have an assigned Version"),
			},
		},
	})
}

func TestAcc_Packer_Data_Artifact_InvalidInputs(t *testing.T) {
	bucketName := testutils.CreateTestSlug("Invalid")
	latestChannel := "latest"
	fingerprint := acctest.RandString(32)
	artifactPlatform := "aws"
	region := "us-east-1"

	// Sets both channel name and version fingerprint
	channelAndFingerprintConfig := packerconfig.NewArtifactDataSourceBuilder("simple")
	channelAndFingerprintConfig.SetBucketName(fmt.Sprintf("%q", bucketName))
	channelAndFingerprintConfig.SetChannelName(fmt.Sprintf("%q", latestChannel))
	channelAndFingerprintConfig.SetVersionFingerprint(fmt.Sprintf("%q", fingerprint))
	channelAndFingerprintConfig.SetPlatform(fmt.Sprintf("%q", artifactPlatform))
	channelAndFingerprintConfig.SetRegion(fmt.Sprintf("%q", region))

	// Missing channel and fingerprint
	missingChannelAndFingerprintConfig := packerconfig.NewArtifactDataSourceBuilder("simple")
	missingChannelAndFingerprintConfig.SetBucketName(fmt.Sprintf("%q", bucketName))
	missingChannelAndFingerprintConfig.SetChannelName(fmt.Sprintf("%q", latestChannel))
	missingChannelAndFingerprintConfig.SetVersionFingerprint(fmt.Sprintf("%q", fingerprint))
	missingChannelAndFingerprintConfig.SetPlatform(fmt.Sprintf("%q", artifactPlatform))
	missingChannelAndFingerprintConfig.SetRegion(fmt.Sprintf("%q", region))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configbuilder.BuildersToString(
					channelAndFingerprintConfig,
				),
				ExpectError: regexp.MustCompile(".*Exactly one of these attributes must be configured.*"),
			},
			{
				Config: configbuilder.BuildersToString(
					missingChannelAndFingerprintConfig,
				),
				ExpectError: regexp.MustCompile(".*Exactly one of these attributes must be configured.*"),
			},
		},
	})
}
