// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package artifact_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder/packerconfig"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testcheck"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func checkDataSource(
	t *testing.T,
	ds packerconfig.ArtifactDataSourceBuilder,
	loc location.BucketLocation,
	versionPtr **packerv2.Version,
	buildPtr **packerv2.Build,
	region string,
) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		var version *packerv2.Version
		if versionPtr != nil {
			version = *versionPtr
		}
		if version == nil {
			version = &packerv2.Version{}
		}

		var build *packerv2.Build
		if buildPtr != nil {
			build = *buildPtr
		}
		if build == nil {
			build = &packerv2.Build{}
		}

		var matchingArtifacts []*packerv2.Artifact
		for _, artifact := range build.Artifacts {
			if artifact.Region == region {
				matchingArtifacts = append(matchingArtifacts, artifact)
			}
		}
		if len(matchingArtifacts) == 0 {
			return fmt.Errorf("didn't find any Artifacts in the Build in the specified region")
		}
		var artifact = matchingArtifacts[0]
		if len(matchingArtifacts) > 1 {
			t.Logf("found %d Artifacts in the provided Build in the specified region, the first will be used", len(matchingArtifacts))
		}

		checks := []resource.TestCheckFunc{
			testcheck.Attribute(ds, "organization_id", loc.GetOrganizationID()),
			testcheck.Attribute(ds, "project_id", loc.GetProjectID()),
			testcheck.Attribute(ds, "bucket_name", loc.GetBucketName()),
			testcheck.Attribute(ds, "platform", build.Platform),
			testcheck.Attribute(ds, "region", artifact.Region),
			testcheck.Attribute(ds, "version_fingerprint", version.Fingerprint),
			testcheck.Attribute(ds, "component_type", build.ComponentType),
			testcheck.Attribute(ds, "id", artifact.ID),
			testcheck.Attribute(ds, "build_id", build.ID),
			testcheck.Attribute(ds, "packer_run_uuid", build.PackerRunUUID),
			testcheck.Attribute(ds, "external_identifier", build.SourceExternalIdentifier),
			testcheck.Attribute(ds, "created_at", artifact.CreatedAt.String()),
		}

		if !version.RevokeAt.IsZero() {
			checks = append(checks, testcheck.Attribute(ds, "revoke_at", version.RevokeAt.String()))
		}

		for key, value := range build.Labels {
			checks = append(checks, testcheck.Attribute(
				ds, fmt.Sprintf("labels.%s", key), value,
			))
		}

		return resource.ComposeAggregateTestCheckFunc(checks...)(state)
	}
}
