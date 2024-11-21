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

var _ rotatingSecret = &awsRotatingSecret{}

type awsRotatingSecret struct{}

func (s *awsRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetAwsIAMUserAccessKeyRotatingSecretConfig(
		secret_service.NewGetAwsIAMUserAccessKeyRotatingSecretConfigParamsWithContext(ctx).
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

func (s *awsRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.AWSAccessKeys == nil {
		return nil, fmt.Errorf("missing required field 'aws_access_keys'")
	}

	response, err := client.CreateAwsIAMUserAccessKeyRotatingSecret(
		secret_service.NewCreateAwsIAMUserAccessKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateAwsIAMUserAccessKeyRotatingSecretBody{
				AwsIamUserAccessKeyParams: &secretmodels.Secrets20231128AwsIAMUserAccessKeyParams{
					Username: secret.AWSAccessKeys.IAMUsername.ValueString(),
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

func (s *awsRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.AWSAccessKeys == nil {
		return nil, fmt.Errorf("missing required field 'aws_access_keys'")
	}

	response, err := client.UpdateAwsIAMUserAccessKeyRotatingSecret(
		secret_service.NewUpdateAwsIAMUserAccessKeyRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateAwsIAMUserAccessKeyRotatingSecretBody{
				AwsIamUserAccessKeyParams: &secretmodels.Secrets20231128AwsIAMUserAccessKeyParams{
					Username: secret.AWSAccessKeys.IAMUsername.ValueString(),
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
