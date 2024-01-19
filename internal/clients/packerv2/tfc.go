package packerv2

import (
	"context"
	"errors"
	"fmt"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// GetRunTask queries the HCP Packer Registry for the API information needed to configure a run task
func GetRunTask(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
) (*packermodels.HashicorpCloudPacker20230101GetRegistryTFCRunTaskAPIResponse, error) {
	params := packerservice.NewPackerServiceGetRegistryTFCRunTaskAPIParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID
	params.TaskType = "validation"

	req, err := client.PackerV2.PackerServiceGetRegistryTFCRunTaskAPI(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceGetRegistryTFCRunTaskAPIDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by GetRunTask. Got: %v", err)
	}

	return req.Payload, nil
}

// RegenerateHMAC triggers the HCP Packer Registry's run task HMAC Key to be regenerated
func RegenerateHMAC(
	ctx context.Context, client *clients.Client, location *sharedmodels.HashicorpCloudLocationLocation,
) (*packermodels.HashicorpCloudPacker20230101RegenerateTFCRunTaskHmacKeyResponse, error) {
	params := packerservice.NewPackerServiceRegenerateTFCRunTaskHmacKeyParamsWithContext(ctx)
	params.LocationOrganizationID = location.OrganizationID
	params.LocationProjectID = location.ProjectID

	req, err := client.PackerV2.PackerServiceRegenerateTFCRunTaskHmacKey(params, nil)
	if err != nil {
		if err, ok := err.(*packerservice.PackerServiceRegenerateTFCRunTaskHmacKeyDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by RegenerateHMAC. Got: %v", err)
	}

	return req.Payload, nil
}
