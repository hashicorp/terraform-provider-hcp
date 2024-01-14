package testclient

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func UpsertVersion(t *testing.T, loc location.BucketLocation, fingerprint string) *packerv2.Version {
	t.Helper()

	client := acctest.HCPClients(t)

	createParams := packerservice.NewPackerServiceCreateVersionParams()
	createParams.SetLocationOrganizationID(loc.GetOrganizationID())
	createParams.SetLocationProjectID(loc.GetProjectID())
	createParams.SetBucketName(loc.GetBucketName())
	createParams.SetBody(&packermodels.HashicorpCloudPacker20230101CreateVersionBody{
		Fingerprint: fingerprint,
	})

	createResp, err := client.PackerV2.PackerServiceCreateVersion(createParams, nil)
	if err == nil {
		// Successful creation
		return createResp.GetPayload().Version
	}

	if err, ok := err.(*packerservice.PackerServiceCreateVersionDefault); !ok || !isAlreadyExistsError(err) {
		// Check if error is because the version already exists, else fail
		t.Errorf("unexpected CreateVersion error during UpsertVersion, expected nil or Already Exists. Got %v", err)
		return nil
	}

	getParams := packerservice.NewPackerServiceGetVersionParams()
	getParams.SetLocationOrganizationID(loc.GetOrganizationID())
	getParams.SetLocationProjectID(loc.GetProjectID())
	getParams.SetBucketName(loc.GetBucketName())
	getParams.SetFingerprint(fingerprint)

	getResp, err := client.PackerV2.PackerServiceGetVersion(getParams, nil)
	if err != nil {
		t.Errorf("unexpected GetVersion error during UpsertVersion, expected nil. Got %v", err)
		return nil
	}

	return getResp.GetPayload().Version
}

func UpsertCompleteVersion(t *testing.T, loc location.BucketLocation, fingerprint string, buildOpts *UpsertBuildOptions) (*packerv2.Version, *packerv2.Build) {
	t.Helper()

	client := acctest.HCPClients(t)

	version := UpsertVersion(t, loc, fingerprint)
	if t.Failed() || version == nil {
		return nil, nil
	}

	options := defaultBuildOptions
	if buildOpts != nil {
		options = *buildOpts
	}
	options.Complete = true

	build := InsertBuild(t,
		&location.GenericVersionLocation{
			BucketLocation:     loc,
			VersionFingerprint: fingerprint,
		},
		&options,
	)
	if t.Failed() {
		return nil, nil
	}

	time.Sleep(15 * time.Second) // Wait a bit to allow the version to be updated fully by cadence

	version, err := packerv2.GetVersionByFingerprint(client, loc, fingerprint)
	if err != nil {
		t.Errorf("version not found after UpsertCompleteVersion, received unexpected error. Got %v", err)
		return nil, nil
	}

	return version, build
}

func RevokeVersion(t *testing.T, loc location.BucketLocation, fingerprint string) *packerv2.Version {
	t.Helper()

	client := acctest.HCPClients(t)

	revokeParams := packerservice.NewPackerServiceUpdateVersionParams()
	revokeParams.SetLocationOrganizationID(loc.GetOrganizationID())
	revokeParams.SetLocationProjectID(loc.GetProjectID())
	revokeParams.SetBucketName(loc.GetBucketName())
	revokeParams.SetFingerprint(fingerprint)
	revokeParams.SetBody(&packermodels.HashicorpCloudPacker20230101UpdateVersionBody{
		RevokeAt: strfmt.DateTime(time.Now()),
	})

	revokeResp, err := client.PackerV2.PackerServiceUpdateVersion(revokeParams, nil)
	if err != nil {
		t.Errorf("unexpected UpdateVersion error during RevokeVersion: %v", err)
	}

	WaitForOperation(t, loc, "Revoke Version", revokeResp.Payload.Operation.ID)
	if t.Failed() {
		return nil
	}

	version, err := packerv2.GetVersionByFingerprint(client, loc, fingerprint)
	if err != nil {
		t.Errorf("version not found after RevokeVersion, received unexpected error: %v", err)
		return nil
	}

	return version
}

func ScheduleRevokeVersion(t *testing.T, loc location.BucketLocation, fingerprint string, revokeAt strfmt.DateTime) *packerv2.Version {
	t.Helper()

	client := acctest.HCPClients(t)

	revokeParams := packerservice.NewPackerServiceUpdateVersionParams()
	revokeParams.SetLocationOrganizationID(loc.GetOrganizationID())
	revokeParams.SetLocationProjectID(loc.GetProjectID())
	revokeParams.SetBucketName(loc.GetBucketName())
	revokeParams.SetFingerprint(fingerprint)
	revokeParams.SetBody(&packermodels.HashicorpCloudPacker20230101UpdateVersionBody{
		RevokeAt: revokeAt,
	})

	revokeResp, err := client.PackerV2.PackerServiceUpdateVersion(revokeParams, nil)
	if err != nil {
		t.Errorf("unexpected UpdateVersion error during ScheduleRevokeVersion, expected nil. Got %v", err)
		return nil
	}

	return revokeResp.GetPayload().Version
}
