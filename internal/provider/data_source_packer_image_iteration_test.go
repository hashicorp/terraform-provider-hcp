package provider

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/google/uuid"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/preview/2020-05-05/client/operation_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"google.golang.org/grpc/codes"
)

const (
	acctestAlpineBucket      = "alpine-acctest"
	acctestUbuntuBucket      = "ubuntu-acctest"
	acctestProductionChannel = "production"
)

var (
	testAccPackerAlpineProductionImage = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}`, acctestAlpineBucket, acctestProductionChannel)
	testAccPackerUbuntuProductionImage = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}`, acctestUbuntuBucket, acctestProductionChannel)
)

func upsertRegistry(t *testing.T) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	params := packer_service.NewPackerServiceCreateRegistryParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.Body = &models.HashicorpCloudPackerCreateRegistryRequest{
		FeatureTier: models.HashicorpCloudPackerRegistryConfigTierSTANDARD,
	}

	resp, err := client.Packer.PackerServiceCreateRegistry(params, nil)
	if err, ok := err.(*packer_service.PackerServiceCreateRegistryDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		default:
			t.Errorf("unexpected CreateRegistry error, expected nil or 409. Got code: %d err: %v", err.Code(), err)
			return
		}
	}

	waitForOperation(t, loc, "Create Registry", resp.Payload.Operation.ID, client)
	return
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

		switch resp.Payload.Operation.State {
		case "PENDING":
			msg := fmt.Sprintf("==> Operation \"%s\" pending...", operationName)
			return fmt.Errorf(msg)
		case "RUNNING":
			msg := fmt.Sprintf("==> Operation \"%s\" running...", operationName)
			return fmt.Errorf(msg)
		case "DONE":
		default:
			t.Errorf("Operation returned unknown state: %s", resp.Payload.Operation.State)
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

func upsertBucket(t *testing.T, bucketSlug string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createBktParams := packer_service.NewPackerServiceCreateBucketParams()
	createBktParams.LocationOrganizationID = loc.OrganizationID
	createBktParams.LocationProjectID = loc.ProjectID
	createBktParams.Body = &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug: bucketSlug,
		Location:   loc,
	}
	_, err := client.Packer.PackerServiceCreateBucket(createBktParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.PackerServiceCreateBucketDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	t.Errorf("unexpected CreateBucket error, expected nil or 409. Got %v", err)
}

func upsertIteration(t *testing.T, bucketSlug, fingerprint string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createItParams := packer_service.NewPackerServiceCreateIterationParams()
	createItParams.LocationOrganizationID = loc.OrganizationID
	createItParams.LocationProjectID = loc.ProjectID
	createItParams.BucketSlug = bucketSlug

	createItParams.Body = &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  bucketSlug,
		Fingerprint: fingerprint,
	}
	_, err := client.Packer.PackerServiceCreateIteration(createItParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.PackerServiceCreateIterationDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	t.Errorf("unexpected CreateIteration error, expected nil or 409. Got %v", err)
}

func revokeIteration(t *testing.T, iterationID, bucketSlug, revokeIn string) {
	t.Helper()
	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	params := packer_service.NewPackerServiceUpdateIterationParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.IterationID = iterationID
	params.Body = &models.HashicorpCloudPackerUpdateIterationRequest{
		BucketSlug: bucketSlug,
		RevokeIn:   revokeIn,
	}

	_, err := client.Packer.PackerServiceUpdateIteration(params, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func getIterationIDFromFingerPrint(t *testing.T, bucketSlug, fingerprint string) string {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = loc.OrganizationID
	getItParams.LocationProjectID = loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.Fingerprint = &fingerprint

	ok, err := client.Packer.PackerServiceGetIteration(getItParams, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ok.Payload.Iteration.ID
}

func upsertBuild(t *testing.T, bucketSlug, fingerprint, iterationID string) {
	client := testAccProvider.Meta().(*clients.Client)

	createBuildParams := packer_service.NewPackerServiceCreateBuildParams()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	createBuildParams.LocationOrganizationID = loc.OrganizationID
	createBuildParams.LocationProjectID = loc.ProjectID
	createBuildParams.BucketSlug = bucketSlug
	createBuildParams.IterationID = iterationID

	createBuildParams.Body = &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug: bucketSlug,
		Build: &models.HashicorpCloudPackerBuildCreateBody{
			CloudProvider: "aws",
			ComponentType: "amazon-ebs.example",
			PackerRunUUID: uuid.New().String(),
			Status:        models.HashicorpCloudPackerBuildStatusRUNNING,
		},
		Fingerprint: fingerprint,
		IterationID: iterationID,
		Location:    loc,
	}

	build, err := client.Packer.PackerServiceCreateBuild(createBuildParams, nil)
	if err, ok := err.(*packer_service.PackerServiceCreateBuildDefault); ok {
		switch err.Code() {
		case int(codes.Aborted), http.StatusConflict:
			// all good here !
			return
		}
	}

	if build == nil {
		t.Errorf("unexpected CreateBuild error, expected non nil build response. Got %v", err)
		return
	}

	// Iterations are currently only assigned an incremental version when publishing image metadata on update.
	// Incremental versions are a requirement for assigning the channel.
	updateBuildParams := packer_service.NewPackerServiceUpdateBuildParams()
	updateBuildParams.LocationOrganizationID = loc.OrganizationID
	updateBuildParams.LocationProjectID = loc.ProjectID
	updateBuildParams.BuildID = build.Payload.Build.ID
	updateBuildParams.Body = &models.HashicorpCloudPackerUpdateBuildRequest{
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			CloudProvider: "aws",
			Status:        models.HashicorpCloudPackerBuildStatusDONE,
			Images: []*models.HashicorpCloudPackerImageCreateBody{
				{
					ImageID: "ami-42",
					Region:  "us-east-1",
				},
				{
					ImageID: "ami-43",
					Region:  "us-east-1",
				},
			},
			Labels: map[string]string{"test-key": "test-value"},
		},
	}
	_, err = client.Packer.PackerServiceUpdateBuild(updateBuildParams, nil)
	if err, ok := err.(*packer_service.PackerServiceUpdateBuildDefault); ok {
		t.Errorf("unexpected UpdateBuild error, expected nil. Got %v", err)
	}
}

func createChannel(t *testing.T, bucketSlug, channelSlug, iterationID string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createChParams := packer_service.NewPackerServiceCreateChannelParams()
	createChParams.LocationOrganizationID = loc.OrganizationID
	createChParams.LocationProjectID = loc.ProjectID
	createChParams.BucketSlug = bucketSlug
	createChParams.Body = &models.HashicorpCloudPackerCreateChannelRequest{
		Slug:        channelSlug,
		IterationID: iterationID,
	}

	_, err := client.Packer.PackerServiceCreateChannel(createChParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.PackerServiceCreateChannelDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			updateChannel(t, bucketSlug, channelSlug, iterationID)
			return
		}
	}
	t.Errorf("unexpected CreateChannel error, expected nil. Got %v", err)
}

func updateChannel(t *testing.T, bucketSlug, channelSlug, iterationID string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	updateChParams := packer_service.NewPackerServiceUpdateChannelParams()
	updateChParams.LocationOrganizationID = loc.OrganizationID
	updateChParams.LocationProjectID = loc.ProjectID
	updateChParams.BucketSlug = bucketSlug
	updateChParams.Slug = channelSlug
	updateChParams.Body = &models.HashicorpCloudPackerUpdateChannelRequest{
		IterationID: iterationID,
	}

	_, err := client.Packer.PackerServiceUpdateChannel(updateChParams, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected UpdateChannel error, expected nil. Got %v", err)
}

func deleteBucket(t *testing.T, bucketSlug string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	deleteBktParams := packer_service.NewPackerServiceDeleteBucketParams()
	deleteBktParams.LocationOrganizationID = loc.OrganizationID
	deleteBktParams.LocationProjectID = loc.ProjectID
	deleteBktParams.BucketSlug = bucketSlug

	_, err := client.Packer.PackerServiceDeleteBucket(deleteBktParams, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected DeleteBucket error, expected nil. Got %v", err)
}

func deleteIteration(t *testing.T, bucketSlug string, iterationID string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	deleteItParams := packer_service.NewPackerServiceDeleteIterationParams()
	deleteItParams.LocationOrganizationID = loc.OrganizationID
	deleteItParams.LocationProjectID = loc.ProjectID
	deleteItParams.BucketSlug = &bucketSlug
	deleteItParams.IterationID = iterationID

	_, err := client.Packer.PackerServiceDeleteIteration(deleteItParams, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected DeleteIteration error, expected nil. Got %v", err)
}

func deleteChannel(t *testing.T, bucketSlug string, channelSlug string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	deleteChParams := packer_service.NewPackerServiceDeleteChannelParams()
	deleteChParams.LocationOrganizationID = loc.OrganizationID
	deleteChParams.LocationProjectID = loc.ProjectID
	deleteChParams.BucketSlug = bucketSlug
	deleteChParams.Slug = channelSlug

	_, err := client.Packer.PackerServiceDeleteChannel(deleteChParams, nil)
	if err == nil {
		return
	}
	t.Errorf("unexpected DeleteChannel error, expected nil. Got %v", err)
}

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.alpine"
	fingerprint := "42"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			itID := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
			deleteChannel(t, acctestAlpineBucket, acctestProductionChannel)
			deleteIteration(t, acctestAlpineBucket, itID)
			deleteBucket(t, acctestAlpineBucket)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestAlpineBucket)
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
					createChannel(t, acctestAlpineBucket, acctestProductionChannel, itID)
				},
				Config: testConfig(testAccPackerAlpineProductionImage),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePacker_revokedIteration(t *testing.T) {
	fingerprint := fmt.Sprintf("%d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			itID := getIterationIDFromFingerPrint(t, acctestUbuntuBucket, fingerprint)
			deleteChannel(t, acctestUbuntuBucket, acctestProductionChannel)
			deleteIteration(t, acctestUbuntuBucket, itID)
			deleteBucket(t, acctestUbuntuBucket)
			return nil
		},
		Steps: []resource.TestStep{
			// testing that getting a revoked iteration fails properly
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestUbuntuBucket)
					upsertIteration(t, acctestUbuntuBucket, fingerprint)
					itID := getIterationIDFromFingerPrint(t, acctestUbuntuBucket, fingerprint)
					upsertBuild(t, acctestUbuntuBucket, fingerprint, itID)
					createChannel(t, acctestUbuntuBucket, acctestProductionChannel, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, acctestUbuntuBucket, "5s")
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config:      testConfig(testAccPackerUbuntuProductionImage),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Error: the iteration (\d|\w){26} is revoked and can not be used`),
			},
		},
	})
}
