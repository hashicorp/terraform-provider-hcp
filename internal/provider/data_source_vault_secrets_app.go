// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceVaultSecretsApp() *schema.Resource {
	return &schema.Resource{
		Description: "The Vault Secrets app data source retrieves secrets and their latest version values for a given application.",
		ReadContext: dataSourceVaultSecretsAppRead,
		Schema: map[string]*schema.Schema{
			"app_name": {
				Description:      "The name of the Vault Secrets application.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"organization_id": {
				Description: "The ID of the HCP organization where the Vault Secrets app is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the Vault Secrets app is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secrets": {
				Description: "A map of all secrets in the Vault Secrets app. Key is the secret name, value is the latest secret version value.",
				Type:        schema.TypeMap,
				Computed:    true,
				Sensitive:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceVaultSecretsAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	appName, ok := d.Get("app_name").(string)
	if !ok {
		return diag.Errorf("Failed to read app_name during data source Read.")
	}
	client := meta.(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	log.Printf("[INFO] Listing Vault Secrets App secrets (%s) [project_id=%s, organization_id=%s]", appName, loc.ProjectID, loc.OrganizationID)

	appSecrets, err := clients.ListVaultSecretsAppSecrets(ctx, client, loc, appName)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("VCS app (%s) does not exist", appName)
		}
		return diag.FromErr(err)
	}

	openAppSecrets := map[string]string{}
	for _, appSecret := range appSecrets {
		secretName := appSecret.Name

		log.Printf("[INFO] Opening latest Vault Secrets App Secret (%s - %s) [project_id=%s, organization_id=%s]", appName, secretName, loc.ProjectID, loc.OrganizationID)

		openSecret, err := clients.OpenVaultSecretsAppSecret(ctx, client, loc, appName, secretName)
		if err != nil {
			if clients.IsResponseCodeNotFound(err) {
				return diag.Errorf("Vault Secrets App Secret (%s - %s) does not exist", appName, secretName)
			}
			return diag.FromErr(err)
		}

		openAppSecrets[secretName] = openSecret.Version.Value
	}

	err = setVaultSecretsAppDataSourceAttributes(d, appName, loc, openAppSecrets)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setVaultSecretsAppDataSourceAttributes(d *schema.ResourceData, appName string, loc *sharedmodels.HashicorpCloudLocationLocation, openSecrets map[string]string) error {
	d.SetId(appName)

	if err := d.Set("organization_id", loc.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}
	if err := d.Set("secrets", openSecrets); err != nil {
		return err
	}

	return nil
}
