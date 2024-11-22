// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ rotatingSecret = &azureRotatingSecret{}

type azureRotatingSecret struct{}

func (s *azureRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetAzureApplicationPasswordRotatingSecretConfig(
		secret_service.NewGetAzureApplicationPasswordRotatingSecretConfigParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Config, nil
}

func (s *azureRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.ConfluentServiceAccount == nil {
		return nil, fmt.Errorf("missing required field 'azure_application_password'")
	}

	response, err := client.CreateAzureApplicationPasswordRotatingSecret(
		secret_service.NewCreateAzureApplicationPasswordRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateAzureApplicationPasswordRotatingSecretBody{
				IntegrationName:    secret.IntegrationName.ValueString(),
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				AzureApplicationPasswordParams: &secretmodels.Secrets20231128AzureApplicationPasswordParams{
					AppClientID: secret.AzureApplicationPassword.AppClientID.ValueString(),
					AppObjectID: secret.AzureApplicationPassword.AppObjectID.ValueString(),
				},
				Name: secret.Name.ValueString(),
			}),
		nil)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Config, nil
}

func (s *azureRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.ConfluentServiceAccount == nil {
		return nil, fmt.Errorf("missing required field 'azure_applicartion_password'")
	}
	response, err := client.UpdateAzureApplicationPasswordRotatingSecret(
		secret_service.NewUpdateAzureApplicationPasswordRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateAzureApplicationPasswordRotatingSecretBody{
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				AzureApplicationPasswordParams: &secretmodels.Secrets20231128AzureApplicationPasswordParams{
					AppClientID: secret.AzureApplicationPassword.AppClientID.ValueString(),
					AppObjectID: secret.AzureApplicationPassword.AppObjectID.ValueString(),
				},
			}),
		nil)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Config, nil
}
