// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testclient

import (
	"testing"

	"github.com/google/uuid"
	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

// Basic options for a build with a single image
var defaultBuildOptions = UpsertBuildOptions{
	Platform:      "aws",
	ComponentType: "amazon-ebs.example." + acctest.RandString(8),
	Artifacts: []*packermodels.HashicorpCloudPacker20230101ArtifactCreateBody{
		{
			ExternalIdentifier: "ami-1234",
			Region:             "us-east-1",
		},
	},
	Complete: true,
}

type UpsertBuildOptions struct {
	Labels        map[string]string
	Platform      string
	ComponentType string
	Artifacts     []*packerv2.CreateArtifactBody
	Complete      bool
}

func InsertBuild(t *testing.T, loc location.VersionLocation, optionsPtr *UpsertBuildOptions) *packerv2.Build {
	// Note: This is insert instead of upsert because it is annoying to find the build ID of an
	// already created build and we shouldn't be running into pre-existing builds very often anyway
	t.Helper()

	var options = defaultBuildOptions
	if optionsPtr != nil {
		options = *optionsPtr
	}

	client := acctest.HCPClients(t)

	createParams := packerservice.NewPackerServiceCreateBuildParams()
	createParams.SetLocationOrganizationID(loc.GetOrganizationID())
	createParams.SetLocationProjectID(loc.GetProjectID())
	createParams.SetBucketName(loc.GetBucketName())
	createParams.SetFingerprint(loc.GetVersionFingerprint())
	createStatus := packermodels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING
	createParams.SetBody(&packermodels.HashicorpCloudPacker20230101CreateBuildBody{
		Platform:      options.Platform,
		ComponentType: options.ComponentType,
		PackerRunUUID: uuid.New().String(),
		Status:        &createStatus,
		Labels:        options.Labels,
	})

	createResp, err := client.PackerV2.PackerServiceCreateBuild(createParams, nil)
	if err != nil {
		t.Fatalf("unexpected CreateBuild error during UpsertBuild: %v", err)
		return nil
	}

	// Creation succeeded
	// If we're not completing the build or adding artifacts, we're done
	if !options.Complete && len(options.Artifacts) == 0 {
		return createResp.GetPayload().Build
	}

	updateParams := packerservice.NewPackerServiceUpdateBuildParams()
	updateParams.SetLocationOrganizationID(loc.GetOrganizationID())
	updateParams.SetLocationProjectID(loc.GetProjectID())
	updateParams.SetBucketName(loc.GetBucketName())
	updateParams.SetFingerprint(loc.GetVersionFingerprint())
	// updateParams.SetLocationRegionProvider(&options.CloudProvider)
	updateParams.SetBuildID(createResp.GetPayload().Build.ID)
	updateStatus := packermodels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET.Pointer()
	if options.Complete {
		updateStatus = packermodels.HashicorpCloudPacker20230101BuildStatusBUILDDONE.Pointer()
	}
	updateParams.SetBody(&packermodels.HashicorpCloudPacker20230101UpdateBuildBody{
		PackerRunUUID: uuid.New().String(),
		Status:        updateStatus,
		Labels:        options.Labels,
		Artifacts:     options.Artifacts,
	})

	updateResp, err := client.PackerV2.PackerServiceUpdateBuild(updateParams, nil)
	if err != nil {
		t.Fatalf("unexpected UpdateBuild error during UpsertBuild: %v", err)
		return nil
	}

	return updateResp.GetPayload().Build
}
