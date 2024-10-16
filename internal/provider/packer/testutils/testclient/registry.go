// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testclient

import (
	"testing"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func UpsertRegistry(t *testing.T, loc location.ProjectLocation, featureTier *packerv2.RegistryFeatureTier) *packerv2.Registry {
	t.Helper()

	client := acctest.HCPClients(t)
	createParams := packerservice.NewPackerServiceCreateRegistryParams()
	createParams.SetLocationOrganizationID(loc.GetOrganizationID())
	createParams.SetLocationProjectID(loc.GetProjectID())

	if featureTier == nil {
		featureTier = packermodels.HashicorpCloudPacker20230101RegistryConfigTierPLUS.Pointer()
	}
	createParams.SetBody(&packermodels.HashicorpCloudPacker20230101CreateRegistryBody{
		FeatureTier: featureTier,
	})

	createResp, err := client.PackerV2.PackerServiceCreateRegistry(createParams, nil)
	if err == nil {
		// Operation started, wait for creation to complete/fail
		WaitForOperation(t, loc, "Create Registry", createResp.Payload.Operation.ID)
		if t.Failed() {
			return nil
		}
		// Don't return, for registry we need to get the current status
		// regardless of whether it is newly created or pre-existing
	} else {
		// Creation failed before operation, check if it's because it already exists
		grpcErr, ok := err.(*packerservice.PackerServiceCreateRegistryDefault)
		if !ok || !isAlreadyExistsError(grpcErr) {
			t.Fatalf("unexpected CreateRegistry error during UpsertRegistry, expected nil or Already Exists. Got: %v", err)
			return nil
		}
	}

	getParams := packerservice.NewPackerServiceGetRegistryParams()
	getParams.SetLocationOrganizationID(loc.GetOrganizationID())
	getParams.SetLocationProjectID(loc.GetProjectID())

	getResp, err := client.PackerV2.PackerServiceGetRegistry(getParams, nil)
	if err != nil {
		t.Fatalf("unexpected GetRegistry error during UpsertRegistry: %v", err)
		return nil
	}

	needsFeatureTierUpdate := false
	// Check to see if current currrentRegistry needs to be updated
	currrentRegistry := getResp.GetPayload().Registry
	if currrentRegistry == nil || currrentRegistry.Config == nil || currrentRegistry.Config.FeatureTier == nil {
		needsFeatureTierUpdate = true
	} else {
		needsFeatureTierUpdate = *currrentRegistry.Config.FeatureTier != *featureTier && *featureTier != packermodels.HashicorpCloudPacker20230101RegistryConfigTierUNSET
	}

	needsReactivation := !getResp.Payload.Registry.Config.Activated

	if !(needsFeatureTierUpdate || needsReactivation) {
		return getResp.GetPayload().Registry
	}

	updateParams := packerservice.NewPackerServiceUpdateRegistryParams()
	updateParams.SetLocationOrganizationID(loc.GetOrganizationID())
	updateParams.SetLocationProjectID(loc.GetProjectID())
	updateParams.SetBody(&packermodels.HashicorpCloudPacker20230101UpdateRegistryBody{})

	if needsFeatureTierUpdate {
		updateParams.Body.FeatureTier = featureTier
		if len(updateParams.Body.UpdateMask) > 0 {
			updateParams.Body.UpdateMask += ","
		}
		updateParams.Body.UpdateMask += "featureTier"
	}
	if needsReactivation {
		updateParams.Body.Activated = true
		if len(updateParams.Body.UpdateMask) > 0 {
			updateParams.Body.UpdateMask += ","
		}
		updateParams.Body.UpdateMask += "activated"
	}

	updateResp, err := client.PackerV2.PackerServiceUpdateRegistry(updateParams, nil)
	if err != nil {
		t.Fatalf("unexpected UpdateRegistry error during UpsertRegistry: %v", err)
		return nil
	}
	WaitForOperation(t, loc, "Update Registry", updateResp.GetPayload().Operation.ID)
	if t.Failed() {
		return nil
	}

	getResp, err = client.PackerV2.PackerServiceGetRegistry(getParams, nil)
	if err != nil {
		t.Fatalf("unexpected GetRegistry error after Update in UpsertRegistry: %v", err)
		return nil
	}

	return getResp.GetPayload().Registry
}
