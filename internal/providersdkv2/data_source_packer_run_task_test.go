// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
)

func TestAcc_Packer_dataSourcePackerRunTask(t *testing.T) {
	runTask := testAccPackerDataRunTaskBuilder("runTask")
	config := testConfig(testAccConfigBuildersToString(runTask))

	// Must not be Parallel, conflicts with test for the equivalent resource
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t, map[string]bool{"aws": false, "azure": false})
			upsertRegistry(t)
		},
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.BlockName()),
			},
			{ // Change HMAC and check that it updates in state
				PreConfig: func() {
					client := testAccProvider.Meta().(*clients.Client)
					loc := &sharedmodels.HashicorpCloudLocationLocation{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}
					_, err := packerv2.RegenerateHMAC(context.Background(), client, loc)
					if err != nil {
						t.Errorf("error while regenerating HMAC key: %v", err)
						return
					}
				},
				Config: config,
				Check:  testAccCheckPackerRunTaskStateMatchesAPI(runTask.BlockName()),
			},
		},
	})
}

func testAccPackerDataRunTaskBuilder(uniqueName string) testAccConfigBuilderInterface {
	return &testAccResourceConfigBuilder{
		isData:       true,
		resourceType: "hcp_packer_run_task",
		uniqueName:   uniqueName,
	}
}
