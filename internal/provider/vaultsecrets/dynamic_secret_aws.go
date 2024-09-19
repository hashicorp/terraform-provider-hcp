// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ dynamicSecret = &awsDynamicSecret{}

type awsDynamicSecret struct{}

func (s *awsDynamicSecret) readFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
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

func (s *awsDynamicSecret) createFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.CreateAwsDynamicSecret(
		secret_service.NewCreateAwsDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateAwsDynamicSecretBody{
				DefaultTTL:      secret.DefaultTtl.ValueString(),
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

func (s *awsDynamicSecret) updateFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.UpdateAwsDynamicSecret(
		secret_service.NewUpdateAwsDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateAwsDynamicSecretBody{
				DefaultTTL: secret.DefaultTtl.ValueString(),
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

func (s *awsDynamicSecret) fromModel(secret *DynamicSecret, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	secretModel, ok := model.(*secretmodels.Secrets20231128AwsDynamicSecret)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128AwsDynamicSecret, got: %T", model))
		return diags
	}

	secret.DefaultTtl = types.StringValue(secretModel.DefaultTTL)
	secret.AWSAssumeRole = &awsAssumeRole{
		IAMRoleARN: types.StringValue(secretModel.AssumeRole.RoleArn),
	}

	return diags
}
