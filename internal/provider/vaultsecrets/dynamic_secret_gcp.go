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

var _ dynamicSecret = &gcpDynamicSecret{}

type gcpDynamicSecret struct{}

func (s *gcpDynamicSecret) readFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.GetGcpDynamicSecret(
		secret_service.NewGetGcpDynamicSecretParamsWithContext(ctx).
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

func (s *gcpDynamicSecret) createFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.CreateGcpDynamicSecret(
		secret_service.NewCreateGcpDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateGcpDynamicSecretBody{
				DefaultTTL:      secret.DefaultTtl.ValueString(),
				IntegrationName: secret.IntegrationName.ValueString(),
				Name:            secret.Name.ValueString(),
				ServiceAccountImpersonation: &secretmodels.Secrets20231128ServiceAccountImpersonationRequest{
					ServiceAccountEmail: secret.GCPImpersonateServiceAccount.ServiceAccountEmail.ValueString(),
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

func (s *gcpDynamicSecret) updateFunc(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error) {
	response, err := client.UpdateGcpDynamicSecret(
		secret_service.NewUpdateGcpDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithName(secret.Name.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateGcpDynamicSecretBody{
				DefaultTTL: secret.DefaultTtl.ValueString(),
				ServiceAccountImpersonation: &secretmodels.Secrets20231128ServiceAccountImpersonationRequest{
					ServiceAccountEmail: secret.GCPImpersonateServiceAccount.ServiceAccountEmail.ValueString(),
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

func (s *gcpDynamicSecret) fromModel(secret *DynamicSecret, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	secretModel, ok := model.(*secretmodels.Secrets20231128GcpDynamicSecret)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128GcpDynamicSecret, got: %T", model))
		return diags
	}

	secret.DefaultTtl = types.StringValue(secretModel.DefaultTTL)
	secret.GCPImpersonateServiceAccount = &gcpImpersonateServiceAccount{
		ServiceAccountEmail: types.StringValue(secretModel.ServiceAccountImpersonation.ServiceAccountEmail),
	}

	return diags
}
