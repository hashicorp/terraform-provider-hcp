package providersdkv2

import (
	"net/http"
	"testing"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"google.golang.org/grpc/codes"
)

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
