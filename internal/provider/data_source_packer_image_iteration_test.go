// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	acctestAlpineBucket      = fmt.Sprintf("alpine-acc-%s", time.Now().Format("200601021504"))
	acctestUbuntuBucket      = fmt.Sprintf("ubuntu-acc-%s", time.Now().Format("200601021504"))
	acctestProductionChannel = fmt.Sprintf("packer-acc-channel-%s", time.Now().Format("200601021504"))
)

var (
	testAccPackerAlpineProductionImage = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}`, acctestAlpineBucket, acctestProductionChannel)
	testAccPackerUbuntuProductionImage = fmt.Sprintf(`
	data "hcp_packer_image_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}`, acctestUbuntuBucket, acctestProductionChannel)
)

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.alpine"
	fingerprint := "42"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestAlpineBucket, acctestProductionChannel, false)
			deleteIteration(t, acctestAlpineBucket, fingerprint, false)
			deleteBucket(t, acctestAlpineBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestAlpineBucket)
					upsertIteration(t, acctestAlpineBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestAlpineBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestAlpineBucket, fingerprint, itID)
					createChannel(t, acctestAlpineBucket, acctestProductionChannel, itID)
				},
				Config: testConfig(testAccPackerAlpineProductionImage),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePacker_revokedIteration(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.ubuntu"
	fingerprint := "42"
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestUbuntuBucket, acctestProductionChannel, false)
			deleteIteration(t, acctestUbuntuBucket, fingerprint, false)
			deleteBucket(t, acctestUbuntuBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestUbuntuBucket)
					upsertIteration(t, acctestUbuntuBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestUbuntuBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestUbuntuBucket, fingerprint, itID)
					createChannel(t, acctestUbuntuBucket, acctestProductionChannel, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, acctestUbuntuBucket, revokeAt)
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config: testConfig(testAccPackerUbuntuProductionImage),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "revoke_at", revokeAt.String()),
				),
			},
		},
	})
}
