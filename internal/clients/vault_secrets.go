// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
)

// CreateVaultSecretsApp will create a Vault Secrets application.
func CreateVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string, description string) (*secretmodels.Secrets20230613App, error) {

	createParams := secret_service.NewCreateAppParams()
	createParams.Context = ctx
	createParams.Body.Name = appName
	createParams.Body.Description = description
	createParams.LocationOrganizationID = loc.OrganizationID
	createParams.LocationProjectID = loc.ProjectID

	createResp, err := client.VaultSecrets.CreateApp(createParams, nil)
	if err != nil {
		return nil, err
	}

	return createResp.Payload.App, nil
}

// GetVaultSecretsApp will read a Vault Secrets application
func GetVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) (*secretmodels.Secrets20230613App, error) {
	getParams := secret_service.NewGetAppParams()
	getParams.Context = ctx
	getParams.Name = appName
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.VaultSecrets.GetApp(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.App, nil
}

// UpdateVaultSecretsApp will update an app's description
func UpdateVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string, description string) (*secretmodels.Secrets20230613App, error) {
	updateParams := secret_service.NewUpdateAppParams()
	updateParams.Context = ctx
	updateParams.Name = appName
	updateParams.Body.Description = description
	updateParams.LocationOrganizationID = loc.OrganizationID
	updateParams.LocationProjectID = loc.ProjectID

	updateResp, err := client.VaultSecrets.UpdateApp(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload.App, nil
}

// ListVaultSecretsAppSecrets will retrieve all app secrets metadata for a Vault Secrets application.
func ListVaultSecretsAppSecrets(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) ([]*secretmodels.Secrets20230613Secret, error) {

	listParams := secret_service.NewListAppSecretsParams()
	listParams.Context = ctx
	listParams.AppName = appName
	listParams.LocationOrganizationID = loc.OrganizationID
	listParams.LocationProjectID = loc.ProjectID

	listResp, err := client.VaultSecrets.ListAppSecrets(listParams, nil)
	if err != nil {
		return nil, err
	}

	return listResp.Payload.Secrets, nil
}

// DeleteVaultSecretsApp will delete a Vault Secrets application.
func DeleteVaultSecretsApp(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) error {

	deleteParams := secret_service.NewDeleteAppParams()
	deleteParams.Context = ctx
	deleteParams.Name = appName
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationProjectID = loc.ProjectID

	_, err := client.VaultSecrets.DeleteApp(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateVaultSecretsAppSecret will create a Vault Secrets application secret.
func CreateVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName, secretValue string) (*secretmodels.Secrets20230613Secret, error) {

	createParams := secret_service.NewCreateAppKVSecretParams()
	createParams.Context = ctx
	createParams.AppName = appName
	createParams.Body.Name = secretName
	createParams.Body.Value = secretValue
	createParams.LocationOrganizationID = loc.OrganizationID
	createParams.LocationProjectID = loc.ProjectID

	createResp, err := client.VaultSecrets.CreateAppKVSecret(createParams, nil)
	if err != nil {
		return nil, err
	}

	return createResp.Payload.Secret, nil
}

// OpenVaultSecretsAppSecret will retrieve the latest secret for a Vault Secrets app, including it's value.
func OpenVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName string) (*secretmodels.Secrets20230613OpenSecret, error) {
	getParams := secret_service.NewOpenAppSecretParams()
	getParams.Context = ctx
	getParams.AppName = appName
	getParams.SecretName = secretName
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	var getResp *secret_service.OpenAppSecretOK
	var err error
	for attempt := 0; attempt < retryCount; attempt++ {
		getResp, err = client.VaultSecrets.OpenAppSecret(getParams, nil)
		if err != nil {
			serviceErr, ok := err.(*secret_service.OpenAppSecretDefault)
			if !ok {
				return nil, err
			}
			if shouldRetryErrorCode(serviceErr.Code(), []int{429}) {
				backOffDuration := getAPIBackoffDuration(serviceErr)
				fmt.Printf("The api rate limit exceeded, attempt: %d trying in %d seconds", (attempt + 1), int64(backOffDuration.Seconds()))
				time.Sleep(backOffDuration)
				continue
			}
			return nil, err
		}
		break
	}
	return getResp.Payload.Secret, nil
}

// DeleteVaultSecretsAppSecret will delete a Vault Secrets application secret.
func DeleteVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName string) error {

	deleteParams := secret_service.NewDeleteAppSecretParams()
	deleteParams.Context = ctx
	deleteParams.AppName = appName
	deleteParams.SecretName = secretName
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationProjectID = loc.ProjectID

	_, err := client.VaultSecrets.DeleteAppSecret(deleteParams, nil)
	if err != nil {
		return err
	}

	return nil
}

func getAPIBackoffDuration(serviceErr *secret_service.OpenAppSecretDefault) time.Duration {
	re := regexp.MustCompile(`try again in (\d+) seconds`)
	match := re.FindStringSubmatch(serviceErr.Error())
	baseDelayUnit := 60
	if len(match) > 1 {
		retryCountInSec, err := strconv.Atoi(match[1])
		if err == nil {
			baseDelayUnit = retryCountInSec
		}
	}
	return time.Duration(baseDelayUnit) * time.Second
}
