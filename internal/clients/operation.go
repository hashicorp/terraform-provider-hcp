// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-operation/stable/2020-05-05/client/operation_service"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// maxConsecutiveWaitErrors is the maximum number of consecutive http response errors
// from the operation wait endpoint.
const maxConsecutiveWaitErrors = 4

const operationWaitTimeout = time.Second * 5

// WaitForOperation will poll the operation wait endpoint until an operation
// is DONE, ctx is canceled, or consecutive errors occur waiting for operation to complete.
func WaitForOperation(ctx context.Context, client *Client, operationName string, loc *sharedmodels.HashicorpCloudLocationLocation, operationID string) error {
	// Construct operation wait params.
	waitTimeout := operationWaitTimeout.String()
	waitParams := operation_service.NewWaitParams()
	waitParams.Context = ctx
	waitParams.ID = operationID
	waitParams.Timeout = &waitTimeout
	waitParams.LocationOrganizationID = loc.OrganizationID
	waitParams.LocationProjectID = loc.ProjectID

	// Start with no consecutive errors.
	consecutiveErrors := 0

	for {
		// Use the function to improve break logic of for loop and case statements.
		shouldBreak, err := func() (bool, error) {
			// Prevent the loop from running faster than our timeout, in the case where an error causes the api to respond early.
			notSoonerThan, cancel := context.WithTimeout(context.Background(), operationWaitTimeout)
			defer cancel()

			log.Printf("[INFO] Waiting for %s operation (%s)", operationName, operationID)
			waitResponse, err := client.Operation.Wait(waitParams, nil)
			if err != nil {
				// Increment consecutive errors - intermittent network errors shouldn't
				// cause a all-out failure when waiting for an operation to complete.
				consecutiveErrors++

				// Terminate wait if the number of consecutive errors has exceeded the threshold.
				if consecutiveErrors >= maxConsecutiveWaitErrors {
					return true, err
				}

				log.Printf("[WARN] Error waiting for %s operation (%s), will retry if possible: %s", operationName, operationID, err.Error())
			} else {
				// Reset consecutive errors after a successful response.
				consecutiveErrors = 0

				log.Printf("[INFO] Received state of %s operation (%s): %s", operationName, operationID, *waitResponse.Payload.Operation.State)
				if *waitResponse.Payload.Operation.State == sharedmodels.HashicorpCloudOperationOperationStateDONE {
					if waitResponse.Payload.Operation.Error != nil {
						err := fmt.Errorf("%s operation (%s) failed [code=%d, message=%s]",
							operationName, waitResponse.Payload.Operation.ID, waitResponse.Payload.Operation.Error.Code, waitResponse.Payload.Operation.Error.Message)
						return true, err
					}

					return true, nil
				}
			}

			// Ensure we don't retry fast if the api responds faster than timeout.
			select {
			case <-ctx.Done():
				return true, fmt.Errorf("context canceled waiting for %s operation (%s) to complete", operationName, operationID)
			case <-notSoonerThan.Done():
				return false, nil
			}
		}()
		if err != nil {
			return err
		}

		// Finish loop checking operation state.
		if shouldBreak {
			break
		}
	}

	return nil
}
