// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ rotatingSecret = &confluentRotatingSecret{}

type confluentRotatingSecret struct{}

func (s *confluentRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetConfluentCloudAPIKeyRotatingSecretConfig(
		secret_service.NewGetConfluentCloudAPIKeyRotatingSecretConfigParamsWithContext(ctx).
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

func (s *confluentRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.ConfluentServiceAccount == nil {
		return nil, fmt.Errorf("missing required field 'confluent_service_account'")
	}

	response, err := client.CreateConfluentCloudAPIKeyRotatingSecret(
		secret_service.NewCreateConfluentCloudAPIKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateConfluentCloudAPIKeyRotatingSecretBody{
				IntegrationName:    secret.IntegrationName.ValueString(),
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				ConfluentCloudAPIKeyParams: &secretmodels.Secrets20231128ConfluentCloudAPIKeyParams{
					ServiceAccountID: secret.ConfluentServiceAccount.ServiceAccountID.ValueString(),
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

func (s *confluentRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.ConfluentServiceAccount == nil {
		return nil, fmt.Errorf("missing required field 'confluent_service_account'")
	}
	response, err := client.UpdateConfluentCloudAPIKeyRotatingSecret(
		secret_service.NewUpdateConfluentCloudAPIKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateConfluentCloudAPIKeyRotatingSecretBody{
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				ConfluentCloudAPIKeyParams: &secretmodels.Secrets20231128ConfluentCloudAPIKeyParams{
					ServiceAccountID: secret.ConfluentServiceAccount.ServiceAccountID.ValueString(),
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
