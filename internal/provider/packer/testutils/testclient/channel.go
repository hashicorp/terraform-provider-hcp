package testclient

import (
	"testing"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func UpsertChannel(t *testing.T, loc location.BucketLocation, name string, versionFingerprint string) *packerv2.Channel {
	t.Helper()

	client := acctest.HCPClients(t)

	createParams := packerservice.NewPackerServiceCreateChannelParams()
	createParams.SetLocationOrganizationID(loc.GetOrganizationID())
	createParams.SetLocationProjectID(loc.GetProjectID())
	createParams.SetBucketName(loc.GetBucketName())
	createParams.SetBody(&packermodels.HashicorpCloudPacker20230101CreateChannelBody{
		Name:               name,
		VersionFingerprint: versionFingerprint,
	})

	resp, err := client.PackerV2.PackerServiceCreateChannel(createParams, nil)
	if err == nil {
		// Creation succeeded
		return resp.GetPayload().Channel
	}
	if createErr, ok := err.(*packerservice.PackerServiceCreateChannelDefault); !ok || !isAlreadyExistsError(createErr) {
		// Check if error is because the channel already exists, else fail
		t.Fatalf("unexpected CreateChannel error during UpsertChannel, expected nil or Already Exists. Got: %v", err)
		return nil
	}

	getParams := packerservice.NewPackerServiceGetChannelParams()
	getParams.SetLocationOrganizationID(loc.GetOrganizationID())
	getParams.SetLocationProjectID(loc.GetProjectID())
	getParams.SetBucketName(loc.GetBucketName())
	getParams.SetChannelName(name)

	_, err = client.PackerV2.PackerServiceGetChannel(getParams, nil)
	if err != nil {
		t.Fatalf("unexpected GetChannel error during UpsertChannel, expected nil. Got: %v", err)
		return nil
	}

	updateParams := packerservice.NewPackerServiceUpdateChannelParams()
	updateParams.SetLocationOrganizationID(loc.GetOrganizationID())
	updateParams.SetLocationProjectID(loc.GetProjectID())
	updateParams.SetBucketName(loc.GetBucketName())
	updateParams.SetChannelName(name)
	updateParams.SetBody(&packermodels.HashicorpCloudPacker20230101UpdateChannelBody{
		VersionFingerprint: versionFingerprint,
		UpdateMask:         "versionFingerprint",
	})

	updateResp, err := client.PackerV2.PackerServiceUpdateChannel(updateParams, nil)
	if err != nil {
		t.Fatalf("unexpected UpdateChannel error during UpsertChannel, expected nil. Got: %v", err)
	}

	return updateResp.GetPayload().Channel
}
