// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version_test

import (
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder/packerconfig"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testcheck"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func checkDataSource(
	ds packerconfig.VersionDataSourceBuilder,
	loc location.BucketLocation,
	channelName string,
	versionFingerprint string,
) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		testcheck.Attribute(ds, "organization_id", loc.GetOrganizationID()),
		testcheck.Attribute(ds, "project_id", loc.GetProjectID()),
		testcheck.Attribute(ds, "bucket_name", loc.GetBucketName()),
		testcheck.Attribute(ds, "fingerprint", versionFingerprint),
	}

	if channelName != "" {
		checks = append(checks, testcheck.Attribute(ds, "channel_name", channelName))
	}

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func checkDataSourceMatchesVersion(ds packerconfig.VersionDataSourceBuilder, versionPtr **packerv2.Version) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		var version *packerv2.Version
		if versionPtr != nil {
			version = *versionPtr
		}
		if version == nil {
			version = &packermodels.HashicorpCloudPacker20230101Version{}
		}

		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(ds.DataSourceName(), "bucket_name", version.BucketName),

			resource.TestCheckResourceAttr(ds.DataSourceName(), "fingerprint", version.Fingerprint),
			resource.TestCheckResourceAttr(ds.DataSourceName(), "name", version.Name),
			resource.TestCheckResourceAttr(ds.DataSourceName(), "id", version.ID),

			resource.TestCheckResourceAttr(ds.DataSourceName(), "author_id", version.AuthorID),

			resource.TestCheckResourceAttr(ds.DataSourceName(), "created_at", version.CreatedAt.String()),
			resource.TestCheckResourceAttr(ds.DataSourceName(), "updated_at", version.UpdatedAt.String()),
			resource.TestCheckResourceAttr(ds.DataSourceName(), "revoke_at", version.RevokeAt.String()),
		}

		return resource.ComposeAggregateTestCheckFunc(checks...)(state)
	}
}
