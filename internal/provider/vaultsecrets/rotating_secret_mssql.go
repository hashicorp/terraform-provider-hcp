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

var _ rotatingSecret = &mssqlRotatingSecret{}

type mssqlRotatingSecret struct{}

func (s *mssqlRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetAppRotatingSecret(
		secret_service.NewGetAppRotatingSecretParamsWithContext(ctx).
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
	return response.Payload, nil
}

func (s *mssqlRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.MssqlUsers == nil {
		return nil, fmt.Errorf("missing required field 'mssql_application_password'")
	}

	response, err := client.CreateAppRotatingSecret(
		secret_service.NewCreateAppRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateAppRotatingSecretBody{
				IntegrationName:    secret.IntegrationName.ValueString(),
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				MssqlUserPasswordDetails: &secretmodels.Secrets20231128MssqlUserPasswordDetails{
					Usernames: secret.MssqlUsers.Usernames,
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
	return response.Payload, nil
}

func (s *mssqlRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.MssqlUsers == nil {
		return nil, fmt.Errorf("missing required field 'mssql_application_password'")
	}
	response, err := client.UpdateAppRotatingSecret(
		secret_service.NewUpdateAppRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateAppRotatingSecretBody{
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				MssqlUserPasswordDetails: &secretmodels.Secrets20231128MssqlUserPasswordDetails{
					Usernames: secret.MssqlUsers.Usernames,
				},
			}),
		nil)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload, nil
}
