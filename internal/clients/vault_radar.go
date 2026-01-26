// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"errors"
	"time"

	dsrs "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/data_source_registration_service"
	ics "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_connection_service"
	iss "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/integration_subscription_service"
	rrs "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/resource_service"
	rsms "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/secret_manager_service"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func OnboardRadarSource(ctx context.Context, client *Client, projectID string, source dsrs.OnboardDataSourceBody) (*dsrs.OnboardDataSourceOK, error) {
	onboardParams := dsrs.NewOnboardDataSourceParamsWithTimeout(2 * time.Minute) // gives datasources with "agent" detector type more time to complete
	onboardParams.Context = ctx
	onboardParams.LocationProjectID = projectID
	onboardParams.Body = source

	onboardResp, err := client.RadarSourceRegistrationService.OnboardDataSource(onboardParams, nil)
	if err != nil {
		return nil, err
	}

	return onboardResp, nil
}

func OnboardRadarSecretManager(ctx context.Context, client *Client, projectID string, source rsms.OnboardSecretManagerBody) (*rsms.OnboardSecretManagerOK, error) {
	onboardParams := rsms.NewOnboardSecretManagerParamsWithTimeout(2 * time.Minute) // gives secret manager more time to complete because an agent is involved.
	onboardParams.Context = ctx
	onboardParams.LocationProjectID = projectID
	onboardParams.Body = source

	onboardResp, err := client.RadarSecretManagerService.OnboardSecretManager(onboardParams, nil)
	if err != nil {
		return nil, err
	}

	return onboardResp, nil
}

func GetRadarSource(ctx context.Context, client *Client, projectID, sourceID string) (*dsrs.GetDataSourceByIDOK, error) {
	getParams := dsrs.NewGetDataSourceByIDParams()
	getParams.Context = ctx
	getParams.ID = sourceID
	getParams.LocationProjectID = projectID

	getResp, err := client.RadarSourceRegistrationService.GetDataSourceByID(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp, nil
}

func GetRadarSecretManager(ctx context.Context, client *Client, projectID, sourceID string) (*rsms.GetSecretManagerByIDOK, error) {
	getParams := rsms.NewGetSecretManagerByIDParams()
	getParams.Context = ctx
	getParams.ID = sourceID
	getParams.LocationProjectID = projectID

	getResp, err := client.RadarSecretManagerService.GetSecretManagerByID(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp, nil
}

func OffboardRadarSource(ctx context.Context, client *Client, projectID, sourceID string) error {
	tflog.SetField(ctx, "radar_source_id", sourceID)

	deleteParams := dsrs.NewOffboardDataSourceParams()
	deleteParams.Context = ctx
	deleteParams.LocationProjectID = projectID
	deleteParams.Body = dsrs.OffboardDataSourceBody{
		ID: sourceID,
	}

	tflog.Trace(ctx, "Initiate radar source offboarding.")
	if _, err := client.RadarSourceRegistrationService.OffboardDataSource(deleteParams, nil); err != nil {
		return err
	}

	return WaitOnOffboardRadarSource(ctx, client, projectID, sourceID)
}

func OffboardRadarSecretManager(ctx context.Context, client *Client, projectID, sourceID string) error {
	tflog.SetField(ctx, "radar_source_id", sourceID)

	deleteParams := rsms.NewOffboardSecretManagerParams()
	deleteParams.Context = ctx
	deleteParams.LocationProjectID = projectID
	deleteParams.Body = rsms.OffboardSecretManagerBody{
		ID: sourceID,
	}

	tflog.Trace(ctx, "Initiate radar secret manager offboarding.")
	if _, err := client.RadarSecretManagerService.OffboardSecretManager(deleteParams, nil); err != nil {
		return err
	}

	return WaitOnOffboardRadarSecretManager(ctx, client, projectID, sourceID)
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

func WaitOnOffboardRadarSecretManager(ctx context.Context, client *Client, projectID, sourceID string) error {
	deletionConfirmation := func() (bool, error) {
		tflog.Trace(ctx, "Confirming radar secret manager deletion.")
		if _, err := GetRadarSecretManager(ctx, client, projectID, sourceID); err != nil {
			if IsResponseCodeNotFound(err) {
				// success, resource not found.
				tflog.Trace(ctx, "Success, radar secret manager deletion confirmed.")
				return true, nil
			}

			tflog.Error(ctx, "Failed to confirm radar secret manager deletion.")
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

func UpdateRadarDataSourceToken(ctx context.Context, client *Client, projectID string, tokenBody dsrs.UpdateDataSourceTokenBody) error {
	params := dsrs.NewUpdateDataSourceTokenParamsWithTimeout(2 * time.Minute)
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = tokenBody

	if _, err := client.RadarSourceRegistrationService.UpdateDataSourceToken(params, nil); err != nil {
		return err
	}

	return nil
}

func UpdateRadarSecretManagerToken(ctx context.Context, client *Client, projectID string, tokenBody rsms.UpdateSecretManagerTokenBody) error {
	params := rsms.NewUpdateSecretManagerTokenParamsWithTimeout(2 * time.Minute)
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = tokenBody

	if _, err := client.RadarSecretManagerService.UpdateSecretManagerToken(params, nil); err != nil {
		return err
	}

	return nil
}

func PatchRadarSecretManagerFeatures(ctx context.Context, client *Client, projectID string, tokenBody rsms.PatchSecretManagerFeaturesBody) error {
	params := rsms.NewPatchSecretManagerFeaturesParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = tokenBody

	if _, err := client.RadarSecretManagerService.PatchSecretManagerFeatures(params, nil); err != nil {
		return err
	}

	return nil
}

func CreateIntegrationConnection(ctx context.Context, client *Client, projectID string, connection ics.CreateIntegrationConnectionBody) (*ics.CreateIntegrationConnectionOK, error) {
	params := ics.NewCreateIntegrationConnectionParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = connection

	resp, err := client.RadarConnectionService.CreateIntegrationConnection(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetIntegrationConnectionByID(ctx context.Context, client *Client, projectID, connectionID string) (*ics.GetIntegrationConnectionByIDOK, error) {
	params := ics.NewGetIntegrationConnectionByIDParams()
	params.Context = ctx
	params.ID = connectionID
	params.LocationProjectID = projectID

	resp, err := client.RadarConnectionService.GetIntegrationConnectionByID(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetIntegrationConnectionByName(ctx context.Context, client *Client, projectID, connectionName string) (*ics.GetIntegrationConnectionByNameOK, error) {
	params := ics.NewGetIntegrationConnectionByNameParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = ics.GetIntegrationConnectionByNameBody{
		Name: connectionName,
	}

	resp, err := client.RadarConnectionService.GetIntegrationConnectionByName(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func DeleteIntegrationConnection(ctx context.Context, client *Client, projectID, connectionID string) error {
	params := ics.NewDeleteIntegrationConnectionParams()
	params.Context = ctx
	params.ID = connectionID
	params.LocationProjectID = projectID

	if _, err := client.RadarConnectionService.DeleteIntegrationConnection(params, nil); err != nil {
		return err
	}

	return nil
}

func UpdateIntegrationConnection(ctx context.Context, client *Client, projectID string, connection ics.UpdateIntegrationConnectionBody) error {
	params := ics.NewUpdateIntegrationConnectionParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = connection

	if _, err := client.RadarConnectionService.UpdateIntegrationConnection(params, nil); err != nil {
		return err
	}

	return nil
}

func CreateIntegrationSubscription(ctx context.Context, client *Client, projectID string, connection iss.CreateIntegrationSubscriptionBody) (*iss.CreateIntegrationSubscriptionOK, error) {
	params := iss.NewCreateIntegrationSubscriptionParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = connection

	resp, err := client.RadarSubscriptionService.CreateIntegrationSubscription(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetIntegrationSubscriptionByID(ctx context.Context, client *Client, projectID, connectionID string) (*iss.GetIntegrationSubscriptionByIDOK, error) {
	params := iss.NewGetIntegrationSubscriptionByIDParams()
	params.Context = ctx
	params.ID = connectionID
	params.LocationProjectID = projectID

	resp, err := client.RadarSubscriptionService.GetIntegrationSubscriptionByID(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetIntegrationSubscriptionByName(ctx context.Context, client *Client, projectID, connectionName string) (*iss.GetIntegrationSubscriptionByNameOK, error) {
	params := iss.NewGetIntegrationSubscriptionByNameParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = iss.GetIntegrationSubscriptionByNameBody{
		Name: connectionName,
	}

	resp, err := client.RadarSubscriptionService.GetIntegrationSubscriptionByName(params, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func DeleteIntegrationSubscription(ctx context.Context, client *Client, projectID, connectionID string) error {
	params := iss.NewDeleteIntegrationSubscriptionParams()
	params.Context = ctx
	params.ID = connectionID
	params.LocationProjectID = projectID

	if _, err := client.RadarSubscriptionService.DeleteIntegrationSubscription(params, nil); err != nil {
		return err
	}

	return nil
}

func UpdateIntegrationSubscription(ctx context.Context, client *Client, projectID string, subscription iss.UpdateIntegrationSubscriptionBody) error {
	params := iss.NewUpdateIntegrationSubscriptionParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = subscription

	if _, err := client.RadarSubscriptionService.UpdateIntegrationSubscription(params, nil); err != nil {
		return err
	}

	return nil
}

func SearchRadarResources(ctx context.Context, client *Client, projectID string, body rrs.SearchResourcesBody) (*rrs.SearchResourcesOK, error) {
	params := rrs.NewSearchResourcesParams()
	params.Context = ctx
	params.LocationProjectID = projectID
	params.Body = body

	res, err := client.RadarResourceService.SearchResources(params, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}
