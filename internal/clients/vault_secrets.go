// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
)

// CreateVaultSecretsApp will create a Vault Secrets application.
func CreateVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string, description string) (*secretmodels.Secrets20231128App, error) {

	createParams := secret_service.NewCreateAppParams()
	createParams.Context = ctx
	createParams.Body.Name = appName
	createParams.Body.Description = description
	createParams.OrganizationID = loc.OrganizationID
	createParams.ProjectID = loc.ProjectID

	createResp, err := client.VaultSecrets.CreateApp(createParams, nil)
	if err != nil {
		return nil, err
	}

	return createResp.Payload.App, nil
}

// GetVaultSecretsApp will read a Vault Secrets application
func GetVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*secretmodels.Secrets20231128App, error) {
	getParams := secret_service.NewGetAppParams()
	getParams.Context = ctx
	getParams.Name = appName
	getParams.OrganizationID = loc.OrganizationID
	getParams.ProjectID = loc.ProjectID

	getResp, err := client.VaultSecrets.GetApp(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.App, nil
}

// UpdateVaultSecretsApp will update an app's description
func UpdateVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string, description string) (*secretmodels.Secrets20231128App, error) {
	updateParams := secret_service.NewUpdateAppParams()
	updateParams.Context = ctx
	updateParams.Name = appName
	updateParams.Body.Description = description
	updateParams.OrganizationID = loc.OrganizationID
	updateParams.ProjectID = loc.ProjectID

	updateResp, err := client.VaultSecrets.UpdateApp(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload.App, nil
}

// DeleteVaultSecretsApp will delete a Vault Secrets application.
func DeleteVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) error {

	deleteParams := secret_service.NewDeleteAppParams()
	deleteParams.Context = ctx
	deleteParams.Name = appName
	deleteParams.OrganizationID = loc.OrganizationID
	deleteParams.ProjectID = loc.ProjectID

	_, err := client.VaultSecrets.DeleteApp(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateVaultSecretsAppSecret will create a Vault Secrets application secret.
func CreateVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName, secretValue string) (*secretmodels.Secrets20231128Secret, error) {

	createParams := secret_service.NewCreateAppKVSecretParams()
	createParams.Context = ctx
	createParams.AppName = appName
	createParams.Body = &secretmodels.SecretServiceCreateAppKVSecretBody{}
	createParams.Body.Name = secretName
	createParams.Body.Value = secretValue
	createParams.OrganizationID = loc.OrganizationID
	createParams.ProjectID = loc.ProjectID

	createResp, err := client.VaultSecrets.CreateAppKVSecret(createParams, nil)
	if err != nil {
		return nil, err
	}

	return createResp.Payload.Secret, nil
}

// DeleteVaultSecretsAppSecret will delete a Vault Secrets application secret.
func DeleteVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName string) error {

	deleteParams := secret_service.NewDeleteAppSecretParams()
	deleteParams.Context = ctx
	deleteParams.AppName = appName
	deleteParams.SecretName = secretName
	deleteParams.OrganizationID = loc.OrganizationID
	deleteParams.ProjectID = loc.ProjectID

	_, err := client.VaultSecrets.DeleteAppSecret(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

func shouldRetryWithSleep(ctx context.Context, err ErrorWithCode, attemptNum int, expectedErrorCodes []int) bool {
	if shouldRetryErrorCode(err.Code(), expectedErrorCodes) {
		backOffDuration := getAPIBackoffDuration(err.Error())
		tflog.Debug(ctx, fmt.Sprintf("error: %q, retrying in %d seconds, attempt: %d", http.StatusText(err.Code()), int64(backOffDuration.Seconds()), attemptNum+1))
		time.Sleep(backOffDuration)
		return true
	}
	return false
}

func getAPIBackoffDuration(serviceErrStr string) time.Duration {
	re := regexp.MustCompile(`try again in (\d+) seconds`)
	match := re.FindStringSubmatch(serviceErrStr)
	backoffSeconds := 60
	if len(match) > 1 {
		backoffSecondsOverride, err := strconv.Atoi(match[1])
		if err == nil {
			backoffSeconds = backoffSecondsOverride
		}
	}
	return time.Duration(backoffSeconds) * time.Second
}
