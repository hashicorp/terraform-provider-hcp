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

var _ rotatingSecret = &gcpRotatingSecret{}

type gcpRotatingSecret struct{}

func (s *gcpRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetGcpServiceAccountKeyRotatingSecretConfig(
		secret_service.NewGetGcpServiceAccountKeyRotatingSecretConfigParamsWithContext(ctx).
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

func (s *gcpRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.GCPServiceAccountKey == nil {
		return nil, fmt.Errorf("missing required field 'gcp_service_account_key'")
	}

	response, err := client.CreateGcpServiceAccountKeyRotatingSecret(
		secret_service.NewCreateGcpServiceAccountKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateGcpServiceAccountKeyRotatingSecretBody{
				GcpServiceAccountKeyParams: &secretmodels.Secrets20231128GcpServiceAccountKeyParams{
					ServiceAccountEmail: secret.GCPServiceAccountKey.ServiceAccountEmail.ValueString(),
				},
				IntegrationName:    secret.IntegrationName.ValueString(),
				Name:               secret.Name.ValueString(),
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
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

func (s *gcpRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.AWSAccessKeys == nil {
		return nil, fmt.Errorf("missing required field 'gcp_service_account_key'")
	}

	response, err := client.UpdateGcpServiceAccountKeyRotatingSecret(
		secret_service.NewUpdateGcpServiceAccountKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateGcpServiceAccountKeyRotatingSecretBody{
				GcpServiceAccountKeyParams: &secretmodels.Secrets20231128GcpServiceAccountKeyParams{
					ServiceAccountEmail: secret.GCPServiceAccountKey.ServiceAccountEmail.ValueString(),
				},
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
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
