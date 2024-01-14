package testclient

import (
	"testing"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func UpsertBucket(t *testing.T, loc location.ProjectLocation, name string) *packerv2.Bucket {
	t.Helper()

	client := acctest.HCPClients(t)

	createParams := packerservice.NewPackerServiceCreateBucketParams()
	createParams.SetLocationOrganizationID(loc.GetOrganizationID())
	createParams.SetLocationProjectID(loc.GetProjectID())
	createParams.SetBody(&packermodels.HashicorpCloudPacker20230101CreateBucketBody{
		Name: name,
	})

	createResp, err := client.PackerV2.PackerServiceCreateBucket(createParams, nil)
	if err == nil {
		// Successful creation
		return createResp.GetPayload().Bucket
	}

	if err, ok := err.(*packerservice.PackerServiceCreateBucketDefault); !ok || !isAlreadyExistsError(err) {
		// Check if error is because the bucket already exists, else fail
		t.Fatalf("unexpected CreateBucket error during UpsertBucket, expected nil or Already Exists. Got: %v", err)
		return nil
	}

	getParams := packerservice.NewPackerServiceGetBucketParams()
	getParams.SetLocationOrganizationID(loc.GetOrganizationID())
	getParams.SetLocationProjectID(loc.GetProjectID())
	getParams.SetBucketName(name)

	getResp, err := client.PackerV2.PackerServiceGetBucket(getParams, nil)
	if err != nil {
		t.Fatalf("unexpected GetBucket error during UpsertBucket: %v", err)
		return nil
	}

	return getResp.GetPayload().Bucket
}

func DeleteBucket(t *testing.T, loc location.ProjectLocation, name string) error {
	t.Helper()

	client := acctest.HCPClients(t)

	deleteParams := packerservice.NewPackerServiceDeleteBucketParams()
	deleteParams.SetLocationOrganizationID(loc.GetOrganizationID())
	deleteParams.SetLocationProjectID(loc.GetProjectID())
	deleteParams.SetBucketName(name)

	_, err := client.PackerV2.PackerServiceDeleteBucket(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil

}
