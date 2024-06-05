// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// OpenVaultSecretsAppSecret will retrieve the latest secret for a Vault Secrets app, including it's value.
func OpenVaultSecretsAppSecret(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName string) (*secretmodels.Secrets20231128OpenSecret, error) {
	getParams := secret_service.NewOpenAppSecretParamsWithContext(ctx).
		WithAppName(appName).
		WithSecretName(secretName).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID)

	var getResp *secret_service.OpenAppSecretOK
	var err error
	for attempt := 0; attempt < retryCount; attempt++ {
		getResp, err = client.VaultSecretsPreview.OpenAppSecret(getParams, nil)
		if err != nil {
			var serviceErr *secret_service.OpenAppSecretDefault
			ok := errors.As(err, &serviceErr)
			if !ok {
				return nil, err
			}

			if shouldRetryErrorCode(serviceErr.Code(), []int{http.StatusTooManyRequests}) {
				backOffDuration := getAPIBackoffDuration(serviceErr.Error())
				tflog.Debug(ctx, fmt.Sprintf("The api rate limit has been exceeded, retrying in %d seconds, attempt: %d", int64(backOffDuration.Seconds()), (attempt+1)))
				time.Sleep(backOffDuration)
				continue
			}
			return nil, err
		}
		break
	}

	if getResp == nil {
		return nil, errors.New("unable to get secret")
	}

	return getResp.GetPayload().Secret, nil
}

func OpenVaultSecretsAppSecrets(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) ([]*secretmodels.Secrets20231128OpenSecret, error) {
	params := secret_service.NewOpenAppSecretsParamsWithContext(ctx).
		WithAppName(appName).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID)

	var secrets *secret_service.OpenAppSecretsOK
	var err error
	for attempt := 0; attempt < retryCount; attempt++ {
		secrets, err = client.VaultSecretsPreview.OpenAppSecrets(params, nil)
		if err != nil {
			var serviceErr *secret_service.OpenAppSecretDefault
			ok := errors.As(err, &serviceErr)
			if !ok {
				return nil, err
			}
			if shouldRetryWithSleep(ctx, serviceErr, attempt, []int{http.StatusTooManyRequests}) {
				continue
			}
			return nil, err
		}
		break
	}

	if secrets == nil {
		return nil, errors.New("unable to get secrets")
	}

	return secrets.GetPayload().Secrets, nil
}

func GetRotatingSecretState(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName, secretName string) (*secretmodels.Secrets20231128RotatingSecretState, error) {
	params := secret_service.NewGetRotatingSecretStateParamsWithContext(ctx).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID).
		WithAppName(appName).
		WithSecretName(secretName)

	resp, err := client.VaultSecretsPreview.GetRotatingSecretState(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload().State, nil
}

// CreateMongoDBAtlasRotationIntegration NOTE: currently just needed for tests
func CreateMongoDBAtlasRotationIntegration(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, integrationName, mongodbAtlasPublicKey, mongodbAtlasPrivateKey string) (*secretmodels.Secrets20231128MongoDBAtlasIntegration, error) {
	body := secret_service.CreateMongoDBAtlasIntegrationBody{
		IntegrationName:      integrationName,
		MongodbAPIPublicKey:  mongodbAtlasPublicKey,
		MongodbAPIPrivateKey: mongodbAtlasPrivateKey,
	}
	params := secret_service.NewCreateMongoDBAtlasIntegrationParamsWithContext(ctx).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID).
		WithBody(body)

	resp, err := client.VaultSecretsPreview.CreateMongoDBAtlasIntegration(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload().Integration, nil
}

// DeleteMongoDBAtlasRotationIntegration NOTE: currently just needed for tests
func DeleteMongoDBAtlasRotationIntegration(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, integrationName string) error {
	params := secret_service.NewDeleteMongoDBAtlasIntegrationParamsWithContext(ctx).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID).
		WithIntegrationName(integrationName)

	_, err := client.VaultSecretsPreview.DeleteMongoDBAtlasIntegration(params, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateMongoDBAtlasRotatingSecret NOTE: currently just needed for tests
func CreateMongoDBAtlasRotatingSecret(
	ctx context.Context,
	client *Client,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	appName string,
	requestBody secret_service.CreateMongoDBAtlasRotatingSecretBody,
) (*secretmodels.Secrets20231128CreateMongoDBAtlasRotatingSecretResponse, error) {
	params := secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(ctx).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID).
		WithAppName(appName).
		WithBody(requestBody)

	resp, err := client.VaultSecretsPreview.CreateMongoDBAtlasRotatingSecret(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload(), nil
}
