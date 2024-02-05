// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version_test

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

func TestAcc_Packer_Data_Version_Simple(t *testing.T) {
	// This is also checked further inside resource.ParallelTest, but we need to
	// check it here because the next like DefaultProjectLocation tries to create the provider
	// client, which it doesn't work in all evirnoments.
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}

	loc := acctest.DefaultProjectLocation(t)

	bucketName := testutils.CreateTestSlug("VersionSimple")
	channelName := bucketName // No need for a different slug
	fingerprint := acctest.RandString(32)

	bucketLoc := location.GenericBucketLocation{
		Location:   loc,
		BucketName: bucketName,
	}

	baseVersionConfig := packerconfig.NewVersionDataSourceBuilder("Simple")
	baseVersionConfig.SetBucketName(fmt.Sprintf("%q", bucketName))

	versionConfigChannel := packerconfig.CloneVersionDataSourceBuilder(baseVersionConfig)
	versionConfigChannel.SetChannelName(fmt.Sprintf("%q", channelName))

	versionConfigLatestChannel := packerconfig.CloneVersionDataSourceBuilder(baseVersionConfig)
	versionConfigLatestChannel.SetChannelName(fmt.Sprintf("%q", "latest"))

	scheduledRevokeFingerprint := acctest.RandString(32)
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(24 * time.Hour))

	var scheduledRevokeVersion *packerv2.Version
	var version *packerv2.Version

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testclient.UpsertRegistry(t, loc, nil)
			testclient.UpsertBucket(t, loc, bucketName)
			version, _ = testclient.UpsertCompleteVersion(t, bucketLoc, fingerprint, nil)
			testclient.UpsertChannel(t, bucketLoc, channelName, version.Fingerprint)
			_, _ = testclient.UpsertCompleteVersion(t, bucketLoc, scheduledRevokeFingerprint, nil)
			scheduledRevokeVersion = testclient.ScheduleRevokeVersion(t, bucketLoc, scheduledRevokeFingerprint, revokeAt)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			loc := acctest.DefaultProjectLocation(t)
			if err := testclient.DeleteBucket(t, loc, bucketName); err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			{ // Check that the data source can get a version by channel name
				Config: configbuilder.BuildersToString(
					versionConfigChannel,
				),
				Check: resource.ComposeTestCheckFunc(
					checkDataSource(versionConfigChannel, bucketLoc, channelName, fingerprint),
					checkDataSourceMatchesVersion(versionConfigChannel, &version),
				),
			},
			{ // Check that the data source can get a to-be-revoked version by channel name
				Config: configbuilder.BuildersToString(
					versionConfigLatestChannel,
				),
				Check: resource.ComposeTestCheckFunc(
					checkDataSource(versionConfigLatestChannel, bucketLoc, "latest", scheduledRevokeFingerprint),
					testcheck.Attribute(versionConfigLatestChannel, "revoke_at", revokeAt.String()),
					checkDataSourceMatchesVersion(versionConfigLatestChannel, &scheduledRevokeVersion),
				),
			},
		},
	})
}

func TestAcc_Packer_Data_Version_InvalidInputs(t *testing.T) {
	bucketName := testutils.CreateTestSlug("InvalidInputs")

	baseVersionConfig := packerconfig.NewVersionDataSourceBuilder("Invalid")
	baseVersionConfig.SetBucketName(fmt.Sprintf("%q", bucketName))

	channelNameTooLong := packerconfig.CloneVersionDataSourceBuilder(baseVersionConfig)
	channelNameTooLong.SetChannelName(fmt.Sprintf("%q", acctest.RandString(37)))

	channelMissing := packerconfig.CloneVersionDataSourceBuilder(baseVersionConfig)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // Check that the data source fails when channel name is too long
				Config: configbuilder.BuildersToString(
					channelNameTooLong,
				),
				ExpectError: regexp.MustCompile(".*must be of length 3-36.*"),
			},
			{ // Check that the data source fails when channel name is not set
				Config: configbuilder.BuildersToString(
					channelMissing,
				),
				ExpectError: regexp.MustCompile(".*The argument \"channel_name\" is required*"),
			},
		},
	})
}
