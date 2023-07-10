// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func TestAcc_dataSourcePackerRunTask(t *testing.T) {
	runTask := testAccPackerDataRunTaskBuilder("runTask")
	config := testConfig(testAccConfigBuildersToString(runTask))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
		},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.ResourceName()),
			},
			{ // Change HMAC and check that it updates in state
				PreConfig: func() {
					client := testAccProvider.Meta().(*clients.Client)
					loc := &sharedmodels.HashicorpCloudLocationLocation{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}
					_, err := clients.RegenerateHMAC(context.Background(), client, loc)
					if err != nil {
						t.Errorf("error while regenerating HMAC key: %v", err)
						return
					}
				},
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.ResourceName()),
			},
		},
	})
}

func testAccPackerDataRunTaskBuilder(uniqueName string) testAccConfigBuilderInterface {
	return &testAccConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_run_task",
		uniqueName:   uniqueName,
	}
}
