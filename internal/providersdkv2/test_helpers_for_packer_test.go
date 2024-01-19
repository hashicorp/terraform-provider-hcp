package providersdkv2

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/google/uuid"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/preview/2020-05-05/client/operation_service"
	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
	"google.golang.org/grpc/codes"
)

func upsertRegistry(t *testing.T) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	params := packerservice.NewPackerServiceCreateRegistryParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	featureTier := packermodels.HashicorpCloudPacker20230101RegistryConfigTierPLUS
	params.Body = &packermodels.HashicorpCloudPacker20230101CreateRegistryBody{
		FeatureTier: &featureTier,
	}

	resp, err := client.PackerV2.PackerServiceCreateRegistry(params, nil)

	if err == nil {
		waitForOperation(t, loc, "Create Registry", resp.Payload.Operation.ID, client)
	}

	if err, ok := err.(*packerservice.PackerServiceCreateRegistryDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			getParams := packerservice.NewPackerServiceGetRegistryParams()
			getParams.LocationOrganizationID = loc.OrganizationID
			getParams.LocationProjectID = loc.ProjectID
			getResp, err := client.PackerV2.PackerServiceGetRegistry(getParams, nil)
			if err != nil {
				t.Errorf("unexpected GetRegistry error: %v", err)
				return
			}
			if *getResp.Payload.Registry.Config.FeatureTier != packermodels.HashicorpCloudPacker20230101RegistryConfigTierPLUS {
				// Make sure is a plus registry
				params := packerservice.NewPackerServiceUpdateRegistryParams()
				params.LocationOrganizationID = loc.OrganizationID
				params.LocationProjectID = loc.ProjectID
				featureTier := packermodels.HashicorpCloudPacker20230101RegistryConfigTierPLUS
				params.Body = &packermodels.HashicorpCloudPacker20230101UpdateRegistryBody{
					FeatureTier: &featureTier,
				}
				resp, err := client.PackerV2.PackerServiceUpdateRegistry(params, nil)
				if err != nil {
					t.Errorf("unexpected UpdateRegistry error: %v", err)
					return
				}
				waitForOperation(t, loc, "Reactivate Registry", resp.Payload.Operation.ID, client)
			}
			return
		default:
			t.Errorf("unexpected CreateRegistry error code, expected nil or 409. Got code: %d err: %v", err.Code(), err)
			return
		}
	}

	t.Errorf("unexpected CreateRegistry error, expected nil. Got: %v", err)
}

func waitForOperation(
	t *testing.T,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	operationName string,
	operationID string,
	client *clients.Client,
) {
	timeout := "5s"
	params := operation_service.NewWaitParams()
	params.ID = operationID
	params.Timeout = &timeout
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID

	operation := func() error {
		resp, err := client.Operation.Wait(params, nil)
		if err != nil {
			t.Errorf("unexpected error %#v", err)
		}

		if resp.Payload.Operation.Error != nil {
			t.Errorf("Operation failed: %s", resp.Payload.Operation.Error.Message)
		}

		switch *resp.Payload.Operation.State {
		case sharedmodels.HashicorpCloudOperationOperationStatePENDING:
			msg := fmt.Sprintf("==> Operation \"%s\" pending...", operationName)
			return fmt.Errorf(msg)
		case sharedmodels.HashicorpCloudOperationOperationStateRUNNING:
			msg := fmt.Sprintf("==> Operation \"%s\" running...", operationName)
			return fmt.Errorf(msg)
		case sharedmodels.HashicorpCloudOperationOperationStateDONE:
		default:
			t.Errorf("Operation returned unknown state: %s", *resp.Payload.Operation.State)
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 10 * time.Second
	bo.RandomizationFactor = 0.5
	bo.Multiplier = 1.5
	bo.MaxInterval = 30 * time.Second
	bo.MaxElapsedTime = 40 * time.Minute
	err := backoff.Retry(operation, bo)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
}

func upsertBucket(t *testing.T, bucketName string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createBktParams := packerservice.NewPackerServiceCreateBucketParams()
	createBktParams.LocationOrganizationID = loc.OrganizationID
	createBktParams.LocationProjectID = loc.ProjectID
	createBktParams.Body = &packermodels.HashicorpCloudPacker20230101CreateBucketBody{
		Name: bucketName,
	}
	_, err := client.PackerV2.PackerServiceCreateBucket(createBktParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packerservice.PackerServiceCreateBucketDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	t.Errorf("unexpected CreateBucket error, expected nil or 409. Got %v", err)
}

func deleteBucket(t *testing.T, bucketName string, logOnError bool) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	deleteBktParams := packerservice.NewPackerServiceDeleteBucketParams()
	deleteBktParams.LocationOrganizationID = loc.OrganizationID
	deleteBktParams.LocationProjectID = loc.ProjectID
	deleteBktParams.BucketName = bucketName

	_, err := client.PackerV2.PackerServiceDeleteBucket(deleteBktParams, nil)
	if err == nil {
		return
	}
	if logOnError {
		t.Logf("unexpected DeleteBucket error, expected nil. Got %v", err)
	}
}

func upsertCompleteVersion(
	t *testing.T, bucketName, fingerprint string, options *build,
) (*packermodels.HashicorpCloudPacker20230101Version, *packermodels.HashicorpCloudPacker20230101Build) {
	version := upsertVersion(t, bucketName, fingerprint)
	if t.Failed() || version == nil {
		return nil, nil
	}
	build := upsertCompleteBuild(t, bucketName, version.Fingerprint, options)
	if t.Failed() {
		return nil, nil
	}

	client := testAccProvider.Meta().(*clients.Client)
	bucketLoc := location.GenericBucketLocation{
		Location: location.GenericLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      client.Config.ProjectID,
		},
		BucketName: bucketName,
	}
	version, err := packerv2.GetVersionByFingerprint(client, bucketLoc, version.Fingerprint)
	if err != nil {
		t.Errorf("Complete version not found after upserting, received unexpected error. Got %v", err)
		return nil, nil
	}

	return version, build
}

func upsertVersion(
	t *testing.T, bucketName, fingerprint string,
) *packermodels.HashicorpCloudPacker20230101Version {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createVersionParams := packerservice.NewPackerServiceCreateVersionParams()
	createVersionParams.LocationOrganizationID = loc.OrganizationID
	createVersionParams.LocationProjectID = loc.ProjectID
	createVersionParams.BucketName = bucketName
	createVersionParams.Body = &packermodels.HashicorpCloudPacker20230101CreateVersionBody{
		Fingerprint: fingerprint,
	}

	versionResp, err := client.PackerV2.PackerServiceCreateVersion(createVersionParams, nil)
	if err == nil {
		return versionResp.Payload.Version
	} else if err, ok := err.(*packerservice.PackerServiceGetVersionDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			getVersionParams := packerservice.NewPackerServiceGetVersionParams()
			getVersionParams.LocationOrganizationID = createVersionParams.LocationOrganizationID
			getVersionParams.LocationProjectID = createVersionParams.LocationProjectID
			getVersionParams.BucketName = createVersionParams.BucketName
			getVersionParams.Fingerprint = createVersionParams.Body.Fingerprint
			versionResp, err := client.PackerV2.PackerServiceGetVersion(getVersionParams, nil)
			if err != nil {
				t.Errorf("unexpected PackerServiceGetVersion error, expected nil. Got %v", err)
				return nil
			}
			return versionResp.Payload.Version
		}
	}

	t.Errorf("unexpected PackerServiceCreateVersion error, expected nil or 409. Got %v", err)
	return nil
}

type build struct {
	labels        map[string]string
	platform      string
	componentType string
	artifacts     []*packermodels.HashicorpCloudPacker20230101ArtifactCreateBody
}

var defaultBuild = build{
	platform:      "aws",
	componentType: "amazon-ebs.example",
	artifacts: []*packermodels.HashicorpCloudPacker20230101ArtifactCreateBody{
		{
			ExternalIdentifier: "ami-1234",
			Region:             "us-east-1",
		},
	},
}

func upsertCompleteBuild(
	t *testing.T, bucketName, fingerprint string, optionsPtr *build,
) *packermodels.HashicorpCloudPacker20230101Build {
	var options = defaultBuild
	if optionsPtr != nil {
		options = *optionsPtr
	}

	client := testAccProvider.Meta().(*clients.Client)

	createBuildParams := packerservice.NewPackerServiceCreateBuildParams()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	createBuildParams.LocationOrganizationID = loc.OrganizationID
	createBuildParams.LocationProjectID = loc.ProjectID
	createBuildParams.BucketName = bucketName
	createBuildParams.Fingerprint = fingerprint

	status := packermodels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING
	createBuildParams.Body = &packermodels.HashicorpCloudPacker20230101CreateBuildBody{
		Platform:      options.platform,
		ComponentType: options.componentType,
		PackerRunUUID: uuid.New().String(),
		Status:        &status,
		Labels:        options.labels,
	}

	var build *packermodels.HashicorpCloudPacker20230101Build

	if createResp, err := client.PackerV2.PackerServiceCreateBuild(createBuildParams, nil); err != nil {
		createErr, ok := err.(*packerservice.PackerServiceCreateBuildDefault)
		if !ok || !(createErr.Code() == int(codes.Aborted) || createErr.Code() == http.StatusConflict) {
			t.Fatalf("unexpected CreateBuild error, expected nil. Got %v", createErr)
		}

		getResp, err := client.PackerV2.PackerServiceGetBuild(&packerservice.PackerServiceGetBuildParams{}, nil)
		if err != nil {
			t.Fatalf("unexpected GetBuild error, expected nil. Got %v", err)
		}
		if getResp == nil {
			t.Fatalf("unexpected GetBuild response, expected non nil. Got nil.")
		} else if getResp.Payload == nil {
			t.Fatalf("unexpected GetBuild response payload, expected non nil. Got nil.")
		}
		build = getResp.Payload.Build
	} else {
		if createResp == nil {
			t.Fatalf("unexpected CreateBuild response, expected non nil. Got nil.")
		} else if createResp.Payload == nil {
			t.Fatalf("unexpected CreateBuild response payload, expected non nil. Got nil.")
		}
		build = createResp.Payload.Build
	}

	// Iterations are currently only assigned an incremental version when publishing image metadata on update.
	// Incremental versions are a requirement for assigning the channel.
	updateBuildParams := packerservice.NewPackerServiceUpdateBuildParams()
	updateBuildParams.LocationOrganizationID = loc.OrganizationID
	updateBuildParams.LocationProjectID = loc.ProjectID
	updateBuildParams.BuildID = build.ID
	updateBuildParams.Fingerprint = fingerprint
	updateBuildParams.BucketName = bucketName

	updatesStatus := packermodels.HashicorpCloudPacker20230101BuildStatusBUILDDONE
	updateBuildParams.Body = &packermodels.HashicorpCloudPacker20230101UpdateBuildBody{
		Status:    &updatesStatus,
		Artifacts: options.artifacts,
	}
	updateResp, err := client.PackerV2.PackerServiceUpdateBuild(updateBuildParams, nil)
	if err != nil {
		t.Errorf("unexpected UpdateBuild error, expected nil. Got %v", err)
	}
	if updateResp == nil {
		t.Fatalf("unexpected UpdateBuild response, expected non nil. Got nil.")
	} else if updateResp.Payload == nil {
		t.Fatalf("unexpected UpdateBuild response payload, expected non nil. Got nil.")
	}
	return updateResp.Payload.Build
}

func createOrUpdateChannel(t *testing.T, bucketName, channelName, fingerprint string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createChParams := packerservice.NewPackerServiceCreateChannelParams()
	createChParams.LocationOrganizationID = loc.OrganizationID
	createChParams.LocationProjectID = loc.ProjectID
	createChParams.BucketName = bucketName
	createChParams.Body = &packermodels.HashicorpCloudPacker20230101CreateChannelBody{
		Name:               channelName,
		VersionFingerprint: fingerprint,
	}

	_, err := client.PackerV2.PackerServiceCreateChannel(createChParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packerservice.PackerServiceCreateChannelDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			updateChannelAssignment(t, bucketName, channelName, fingerprint)
			return
		}
	}
	t.Errorf("unexpected CreateChannel error, expected nil. Got %v", err)
}

func updateChannelAssignment(t *testing.T, bucketName, channelName, versionFingerprint string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	params := packerservice.NewPackerServiceUpdateChannelParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName
	params.Body = &packermodels.HashicorpCloudPacker20230101UpdateChannelBody{
		VersionFingerprint: versionFingerprint,
		UpdateMask:         "versionFingerprint",
	}

	_, err := client.PackerV2.PackerServiceUpdateChannel(params, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected UpdateChannel error, expected nil. Got %v", err)
}

func updateChannelRestriction(t *testing.T, bucketName string, channelName string, restricted bool) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	params := packerservice.NewPackerServiceUpdateChannelParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName
	params.Body = &packermodels.HashicorpCloudPacker20230101UpdateChannelBody{
		Restricted: restricted,
		UpdateMask: "restricted",
	}

	_, err := client.PackerV2.PackerServiceUpdateChannel(params, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected UpdateChannel error, expected nil. Got %v", err)
}
