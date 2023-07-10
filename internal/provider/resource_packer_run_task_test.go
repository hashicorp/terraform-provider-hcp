// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func TestAccPackerRunTask(t *testing.T) {
	runTask := testAccPackerRunTaskBuilder("runTask", `false`)
	config := testConfig(testAccConfigBuildersToString(runTask))
	runTaskRegen := testAccPackerRunTaskBuilderFromRunTask(runTask, `true`)
	configRegen := testConfig(testAccConfigBuildersToString(runTaskRegen))

	getHmacBeforeStep := func(hmacPtr *string) func() {
		return func() {
			client := testAccProvider.Meta().(*clients.Client)
			loc := &sharedmodels.HashicorpCloudLocationLocation{
				OrganizationID: client.Config.OrganizationID,
				ProjectID:      client.Config.ProjectID,
			}
			resp, err := clients.GetRunTask(context.Background(), client, loc)
			if err != nil {
				t.Errorf("failed to get run task before test step, received error: %v", err)
				return
			}
			*hmacPtr = resp.HmacKey
		}
	}

	var preStep2HmacKey string
	var preStep3HmacKey string
	var preStep4HmacKey string

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
			{ // Ensure HMAC key is different after apply
				PreConfig: getHmacBeforeStep(&preStep2HmacKey),
				Config:    configRegen,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackerRunTaskStateMatchesAPI(runTaskRegen.ResourceName()),
					testAccCheckResourceAttrPtrDifferent(runTaskRegen.ResourceName(), "hmac_key", &preStep2HmacKey),
				),
				ExpectNonEmptyPlan: true, // `regenerate_hmac = true` creates a perpetual diff
			},
			{ // Ensure that repetitive applies without changes still regenerate the HMAC key
				PreConfig: getHmacBeforeStep(&preStep3HmacKey),
				Config:    configRegen,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackerRunTaskStateMatchesAPI(runTaskRegen.ResourceName()),
					testAccCheckResourceAttrPtrDifferent(runTaskRegen.ResourceName(), "hmac_key", &preStep3HmacKey),
				),
				ExpectNonEmptyPlan: true, // `regenerate_hmac = true` creates a perpetual diff
			},
			{ // Ensure that applies with regeneration off don't regenerate
				PreConfig: getHmacBeforeStep(&preStep4HmacKey),
				Config:    config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackerRunTaskStateMatchesAPI(runTaskRegen.ResourceName()),
					resource.TestCheckResourceAttrPtr(runTask.ResourceName(), "hmac_key", &preStep4HmacKey),
				),
			},
		},
	})
}

func testAccPackerRunTaskBuilder(uniqueName string, regenerateHmac string) testAccConfigBuilderInterface {
	return &testAccConfigBuilder{
		resourceType: "hcp_packer_run_task",
		uniqueName:   uniqueName,
		attributes: map[string]string{
			"regenerate_hmac": regenerateHmac,
		},
	}
}

func testAccPackerRunTaskBuilderFromRunTask(oldRT testAccConfigBuilderInterface, regenerateHmac string) testAccConfigBuilderInterface {
	return testAccPackerRunTaskBuilder(
		oldRT.UniqueName(),
		regenerateHmac,
	)
}

func testAccPullPackerRunTaskFromAPIWithRunTaskState(resourceName string, state *terraform.State) (*models.HashicorpCloudPackerGetRegistryTFCRunTaskAPIResponse, error) {
	client := testAccProvider.Meta().(*clients.Client)

	loc, _ := testAccGetLocationFromState(resourceName, state)

	resp, err := clients.GetRunTask(context.Background(), client, loc)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func testAccCheckPackerRunTaskStateMatchesRunTask(resourceName string, runTaskPtr **models.HashicorpCloudPackerGetRegistryTFCRunTaskAPIResponse) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		var runTask *models.HashicorpCloudPackerGetRegistryTFCRunTaskAPIResponse
		if runTaskPtr != nil {
			runTask = *runTaskPtr
		}
		if runTask == nil {
			runTask = &models.HashicorpCloudPackerGetRegistryTFCRunTaskAPIResponse{}
		}

		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "endpoint_url", runTask.APIURL),
			resource.TestCheckResourceAttr(resourceName, "hmac_key", runTask.HmacKey),
		)(state)
	}
}

func testAccCheckPackerRunTaskStateMatchesAPI(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		runTask, err := testAccPullPackerRunTaskFromAPIWithRunTaskState(resourceName, state)
		if err != nil {
			return err
		}

		return testAccCheckPackerRunTaskStateMatchesRunTask(resourceName, &runTask)(state)
	}
}
