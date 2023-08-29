// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/consul"
)

// defaultConsulVersionsTimeoutDuration is the default timeout
// for reading the agent Helm config.
var defaultConsulVersionsTimeoutDuration = time.Minute * 5

// dataSourceConsulVersions is the data source for the Consul versions supported by HCP.
func dataSourceConsulVersions() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul versions data source provides the Consul versions supported by HCP.",
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultConsulVersionsTimeoutDuration,
		},
		ReadContext: dataSourceConsulVersionsRead,
		Schema: map[string]*schema.Schema{
			// Computed outputs
			"recommended": {
				Description: "The recommended Consul version for HCP clusters.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"available": {
				Description: "The Consul versions available on HCP.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"preview": {
				Description: "The preview versions of Consul available on HCP.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

// dataSourceConsulVersionsRead is the func to implement reading of the
// supported Consul versions on HCP.
func dataSourceConsulVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	availableConsulVersions, err := clients.GetAvailableHCPConsulVersions(ctx, meta.(*clients.Client))
	if err != nil || len(availableConsulVersions) == 0 {
		return diag.Errorf("error fetching available HCP Consul versions: %v", err)
	}

	var recommendedVersion string
	availableVersions := make([]string, 0)
	previewVersions := make([]string, 0)

	for _, v := range availableConsulVersions {
		switch *v.Status {
		case consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED:
			recommendedVersion = v.Version
			availableVersions = append(availableVersions, v.Version)
		case consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE:
			availableVersions = append(availableVersions, v.Version)
		case consulmodels.HashicorpCloudConsul20210204VersionStatusPREVIEW:
			previewVersions = append(previewVersions, v.Version)
		}
	}

	err = d.Set("recommended", recommendedVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("available", availableVersions)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("preview", previewVersions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%x", md5.Sum([]byte(consul.VersionsToString(availableConsulVersions)))))

	return nil
}
