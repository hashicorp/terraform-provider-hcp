// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultRootTokenTimeoutDuration is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultRootTokenTimeoutDuration = time.Minute * 5

// rootTokenKubernetesSecretTemplate is the template used to generate a
// kubernetes formatted secret for the cluster root token.
const rootTokenKubernetesSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: %s-bootstrap-token
type: Opaque
data:
  token: %s`

// resourceConsulClusterRootToken represents an HCP Consul cluster.
func resourceConsulClusterRootToken() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster root token resource is the token used to bootstrap the cluster's ACL system. " +
			"You can also generate this root token from the HCP Consul UI.",
		CreateContext: resourceConsulClusterRootTokenCreate,
		ReadContext:   resourceConsulClusterRootTokenRead,
		DeleteContext: resourceConsulClusterRootTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultRootTokenTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugIDOrID,
			},
			// Computed outputs
			"accessor_id": {
				Description: "The accessor ID of the root ACL token.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret_id": {
				Description: "The secret ID of the root ACL token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"kubernetes_secret": {
				Description: "The root ACL token Base64 encoded in a Kubernetes secret.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// resourceClusterRootTokenCreate will generate a new root token for the cluster
func resourceConsulClusterRootTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	if matchesID(clusterID) {
		clusterID = filepath.Base(clusterID)
	}

	// fetch organizationID by project ID
	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	_, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to create root ACL token; Consul cluster (%s) not found",
				clusterID,
			)

		}

		return diag.Errorf("unable to check for presence of an existing Consul cluster (%s): %v",
			clusterID,
			err,
		)
	}

	rootTokenResp, err := clients.CreateCustomerRootACLToken(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("error creating HCP Consul cluster root ACL token (cluster_id %q) (project_id %q): %+v",
			clusterID,
			projectID,
			err,
		)
	}

	// Set root token resource data here since 'read' is a no-op
	err = d.Set("accessor_id", rootTokenResp.ACLToken.AccessorID)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID := rootTokenResp.ACLToken.SecretID
	err = d.Set("secret_id", secretID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("kubernetes_secret", generateKubernetesSecret(secretID, clusterID))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rootTokenResp.ACLToken.AccessorID)

	return nil
}

// resourceConsulClusterRootTokenRead will act as a no-op as the root token is not persisted in
// any way that it can be fetched and read
func resourceConsulClusterRootTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	if matchesID(clusterID) {
		clusterID = filepath.Base(clusterID)
	}
	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	_, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// No cluster exists, so this root token should be removed from state
			log.Printf("[WARN] no HCP Consul cluster found with (cluster_id %q) (project_id %q); removing root token.",
				clusterID,
				projectID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing Consul cluster (cluster_id %q) (project_id %q): %v",
			clusterID,
			projectID,
			err,
		)
	}

	return nil
}

// resourceClusterRootTokenDelete will "delete" an existing token by creating a new one,
// that will not be returned, and invalidating the previous token for the cluster.
func resourceConsulClusterRootTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	if matchesID(clusterID) {
		clusterID = filepath.Base(clusterID)
	}
	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	_, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// No cluster exists, so this root token should be removed from state
			log.Printf("[WARN] no HCP Consul cluster found with (cluster_id %q) (project_id %q); removing root token.",
				clusterID,
				projectID,
			)
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing Consul cluster (%s): %v",
			clusterID,
			err,
		)
	}

	// generate a new token to invalidate the previous one, but discard the response
	_, err = clients.CreateCustomerRootACLToken(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to delete Consul cluster (%s) root ACL token: %v",
			clusterID,
			err,
		)
	}

	return nil
}

// generateKubernetesSecret will generate a Kubernetes secret with
// a base64 encoded root token secret as it's token.
func generateKubernetesSecret(rootTokenSecretID, clusterID string) string {
	return fmt.Sprintf(rootTokenKubernetesSecretTemplate,
		// lowercase the name
		strings.ToLower(clusterID),
		// base64 encode the secret value
		base64.StdEncoding.EncodeToString([]byte(rootTokenSecretID)))
}
