package provider

import (
	"net/http"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"google.golang.org/grpc/codes"
)

var (
	testAccPackerAlpineProductionImage = `
	data "hcp_packer_image_iteration" "alpine" {
		bucket  = "alpine"
		channel = "production"
	}`
)

func upsertBucket(t *testing.T, bucketSlug string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	createBktParams := packer_service.NewCreateBucketParams()
	createBktParams.LocationOrganizationID = loc.OrganizationID
	createBktParams.LocationProjectID = loc.ProjectID
	createBktParams.Body = &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug: bucketSlug,
		Location:   loc,
	}
	_, err := client.Packer.CreateBucket(createBktParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.CreateBucketDefault); ok {
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

	createItParams := packer_service.NewCreateIterationParams()
	createItParams.LocationOrganizationID = loc.OrganizationID
	createItParams.LocationProjectID = loc.ProjectID
	createItParams.BucketSlug = bucketSlug

	createItParams.Body = &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  bucketSlug,
		Fingerprint: fingerprint,
	}
	_, err := client.Packer.CreateIteration(createItParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.CreateIterationDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	t.Errorf("unexpected CreateIteration error, expected nil or 409. Got %v", err)
}

func getIterationIDFromFingerPrint(t *testing.T, bucketSlug, fingerprint string) string {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	getItParams := packer_service.NewGetIterationParams()
	getItParams.LocationOrganizationID = loc.OrganizationID
	getItParams.LocationProjectID = loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.Fingerprint = &fingerprint

	ok, err := client.Packer.GetIteration(getItParams, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ok.Payload.Iteration.ID
}

func upsertBuild(t *testing.T, bucketSlug, fingerprint, iterationID string) {
	client := testAccProvider.Meta().(*clients.Client)

	createBuildParams := packer_service.NewCreateBuildParams()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	createBuildParams.LocationOrganizationID = loc.OrganizationID
	createBuildParams.LocationProjectID = loc.ProjectID
	createBuildParams.BucketSlug = bucketSlug
	createBuildParams.BuildIterationID = iterationID

	createBuildParams.Body = &models.HashicorpCloudPackerCreateBuildRequest{
		Fingerprint: fingerprint,
		BucketSlug:  bucketSlug,
		Location:    loc,
	}
	createBuildParams.Body.Build = &models.HashicorpCloudPackerBuild{
		CloudProvider: "aws",
		ComponentType: "amazon-ebs.example",
		IterationID:   iterationID,
		Status:        models.HashicorpCloudPackerBuildStatusDONE,
		Images: []*models.HashicorpCloudPackerImage{
			{
				ImageID: "ami-42",
			},
			{
				ImageID: "ami-43",
			},
		},
	}
	_, err := client.Packer.CreateBuild(createBuildParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.CreateBuildDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	t.Errorf("unexpected CreateBuild error, expected nil or 409. Got %v", err)
}

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.alpine"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.

			{
				PreConfig: func() {
					bucket := "alpine"
					fingerprint := "42"
					upsertBucket(t, bucket)
					upsertIteration(t, bucket, fingerprint)
					itID := getIterationIDFromFingerPrint(t, bucket, fingerprint)
					upsertBuild(t, bucket, fingerprint, itID)
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
