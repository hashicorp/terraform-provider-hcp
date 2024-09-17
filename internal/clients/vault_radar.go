package clients

import (
	"context"
	"errors"
	"time"

	radar_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/data_source_registration_service"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func OnboardRadarSource(ctx context.Context, client *Client, projectID string, source radar_service.OnboardDataSourceBody) (*radar_service.OnboardDataSourceOK, error) {
	onboardParams := radar_service.NewOnboardDataSourceParams()
	onboardParams.Context = ctx
	onboardParams.LocationProjectID = projectID
	onboardParams.Body = source

	onboardResp, err := client.RadarSourceRegistrationService.OnboardDataSource(onboardParams, nil)
	if err != nil {
		return nil, err
	}

	return onboardResp, nil
}

func GetRadarSource(ctx context.Context, client *Client, projectID, sourceID string) (*radar_service.GetDataSourceByIDOK, error) {
	getParams := radar_service.NewGetDataSourceByIDParams()
	getParams.Context = ctx
	getParams.ID = sourceID
	getParams.LocationProjectID = projectID

	getResp, err := client.RadarSourceRegistrationService.GetDataSourceByID(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp, nil
}

func OffboardRadarSource(ctx context.Context, client *Client, projectID, sourceID string) error {
	tflog.SetField(ctx, "radar_source_id", sourceID)

	deleteParams := radar_service.NewOffboardDataSourceParams()
	deleteParams.Context = ctx
	deleteParams.LocationProjectID = projectID
	deleteParams.Body = radar_service.OffboardDataSourceBody{
		ID: sourceID,
	}

	tflog.Trace(ctx, "Initiate radar source offboarding.")
	if _, err := client.RadarSourceRegistrationService.OffboardDataSource(deleteParams, nil); err != nil {
		return err
	}

	return WaitOnOffboardRadarSource(ctx, client, projectID, sourceID)
}

func WaitOnOffboardRadarSource(ctx context.Context, client *Client, projectID, sourceID string) error {
	deletionConfirmation := func() (bool, error) {
		tflog.Trace(ctx, "Confirming radar source deletion.")
		if _, err := GetRadarSource(ctx, client, projectID, sourceID); err != nil {
			if IsResponseCodeNotFound(err) {
				// success, resource not found.
				tflog.Trace(ctx, "Success, radar source deletion confirmed.")
				return true, nil
			}

			tflog.Error(ctx, "Failed to confirm radar source deletion.")
			return false, err
		}

		// Resource still exists.
		return false, nil
	}

	retry := 10 * time.Second
	timeout := 10 * time.Minute
	maxConsecutiveErrors := 5
	return waitFor(ctx, retry, timeout, maxConsecutiveErrors, deletionConfirmation)
}

// waitFor waits for isDone to return true or retrying every retry duration until timeout.
// Returns an error if isDone errors or timeout expires.
func waitFor(ctx context.Context, retry, timeout time.Duration, maxConsecutiveErrors int, isDone func() (bool, error)) error {
	consecutiveErrors := 0

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(retry)
	defer ticker.Stop()

	for {
		if done, err := isDone(); err != nil {
			// Check for consecutiveErrors and return error if it exceeds the limit.
			if consecutiveErrors >= maxConsecutiveErrors {
				return errors.New("max consecutive errors reached")
			}
			consecutiveErrors++
			// Don't call continue here, as we want to wait for the next retry duration.
		} else if done {
			return nil
		} else {
			// done == false, err == nil
			// Reset consecutiveErrors for next retry call to isDone.
			consecutiveErrors = 0
		}

		select {
		case <-ticker.C:
			// retry duration has passed.
		case <-waitCtx.Done():
			return errors.New("timeout expired while waiting")
		case <-ctx.Done():
			return errors.New("context canceled while waiting")
		}
	}
}
