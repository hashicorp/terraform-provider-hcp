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

var _ dynamicSecret = &awsDynamicSecret{}

type awsDynamicSecret struct{}

func (s *awsDynamicSecret) read(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.GetAwsDynamicSecret(
		secret_service.NewGetAwsDynamicSecretParamsWithContext(ctx).
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
	return response.Payload.Secret, nil
}

func (s *awsDynamicSecret) create(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	if secret.AWSAssumeRole == nil {
		return nil, fmt.Errorf("missing required field 'aws_assume_role'")
	}

	response, err := client.CreateAwsDynamicSecret(
		secret_service.NewCreateAwsDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateAwsDynamicSecretBody{
				DefaultTTL:      secret.DefaultTTL.ValueString(),
				IntegrationName: secret.IntegrationName.ValueString(),
				Name:            secret.Name.ValueString(),
				AssumeRole: &secretmodels.Secrets20231128AssumeRoleRequest{
					RoleArn: secret.AWSAssumeRole.IAMRoleARN.ValueString(),
				},
			}),
		nil)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Secret, nil
}

func (s *awsDynamicSecret) update(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	if secret.AWSAssumeRole == nil {
		return nil, fmt.Errorf("missing required field 'aws_assume_role'")
	}

	response, err := client.UpdateAwsDynamicSecret(
		secret_service.NewUpdateAwsDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateAwsDynamicSecretBody{
				DefaultTTL: secret.DefaultTTL.ValueString(),
				AssumeRole: &secretmodels.Secrets20231128AssumeRoleRequest{
					RoleArn: secret.AWSAssumeRole.IAMRoleARN.ValueString(),
				},
			}),
		nil)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Secret, nil
}
