package testclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/preview/2020-05-05/client/operation_service"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func WaitForOperation(
	t *testing.T,
	loc location.ProjectLocation,
	operationName string,
	operationID string,
) {
	t.Helper()

	timeout := "5s"
	params := operation_service.NewWaitParams()
	params.ID = operationID
	params.Timeout = &timeout
	params.LocationOrganizationID = loc.GetOrganizationID()
	params.LocationProjectID = loc.GetProjectID()

	operation := func() error {
		resp, err := acctest.HCPClients(t).Operation.Wait(params, nil)
		if err != nil {
			t.Errorf("unexpected error %#v", err)
			return nil
		}

		if resp.Payload.Operation.Error != nil {
			t.Errorf("Operation failed: %s", resp.Payload.Operation.Error.Message)
			return nil
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
			return nil
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
