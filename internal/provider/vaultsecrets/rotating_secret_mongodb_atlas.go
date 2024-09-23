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

var _ rotatingSecret = &mongoDBAtlasRotatingSecret{}

type mongoDBAtlasRotatingSecret struct{}

func (s *mongoDBAtlasRotatingSecret) read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.GetMongoDBAtlasRotatingSecretConfig(
		secret_service.NewGetMongoDBAtlasRotatingSecretConfigParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithSecretName(secret.Name.ValueString()), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		return nil, err
	}
	if response == nil || response.Payload == nil {
		return nil, nil
	}
	return response.Payload.Config, nil
}

func (s *mongoDBAtlasRotatingSecret) create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	if secret.MongoDBAtlasUser == nil {
		return nil, fmt.Errorf("missing required field 'mongodb_atlas_user'")
	}

	response, err := client.CreateMongoDBAtlasRotatingSecret(
		secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceCreateMongoDBAtlasRotatingSecretBody{
				IntegrationName:    secret.IntegrationName.ValueString(),
				MongodbGroupID:     secret.MongoDBAtlasUser.ProjectID.ValueString(), // Group ID must be at this level, not in the secret details
				RotationPolicyName: secret.RotationPolicyName.ValueString(),
				SecretDetails: &secretmodels.Secrets20231128MongoDBAtlasSecretDetails{
					MongodbRoles: secret.mongoDBRoles,
				},
				SecretName: secret.Name.ValueString(),
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

func (s *mongoDBAtlasRotatingSecret) update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error) {
	response, err := client.UpdateTwilioRotatingSecret(
		secret_service.NewUpdateTwilioRotatingSecretParamsWithContext(ctx).
			WithOrganizationID(secret.OrganizationID.ValueString()).
			WithProjectID(secret.ProjectID.ValueString()).
			WithAppName(secret.AppName.ValueString()).
			WithBody(&secretmodels.SecretServiceUpdateTwilioRotatingSecretBody{
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
