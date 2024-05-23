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

// ListVaultSecretsAppSecrets will retrieve all app secrets metadata for a Vault Secrets application.
func ListVaultSecretsAppSecrets(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, appName string) ([]*secretmodels.Secrets20231128Secret, error) {
	listParams := secret_service.NewListAppSecretsParamsWithContext(ctx).
		WithAppName(appName).
		WithOrganizationID(loc.OrganizationID).
		WithProjectID(loc.ProjectID)

	listResp, err := client.VaultSecretsPreview.ListAppSecrets(listParams, nil)
	if err != nil {
		return nil, err
	}
	return listResp.GetPayload().Secrets, nil
}

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
