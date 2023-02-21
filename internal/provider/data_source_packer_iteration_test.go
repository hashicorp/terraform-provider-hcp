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
	acctestIterationBucket       = fmt.Sprintf("alpine-acc-itertest-%s", time.Now().Format("200601021504"))
	acctestIterationUbuntuBucket = fmt.Sprintf("ubuntu-acc-itertest-%s", time.Now().Format("200601021504"))
	acctestIterationChannel      = "production-iter-test"
)

var (
	testAccPackerIterationAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}
    # we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine.revoke_at
	}
`, acctestIterationBucket, acctestIterationChannel)

	testAccPackerIterationUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}
`, acctestIterationUbuntuBucket, acctestIterationChannel)
)

func TestAcc_dataSourcePackerIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.alpine"
	fingerprint := "43"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestIterationBucket, acctestIterationChannel, false)
			deleteIteration(t, acctestIterationBucket, fingerprint, false)
			deleteBucket(t, acctestIterationBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, acctestIterationBucket)
					upsertIteration(t, acctestIterationBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestIterationBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestIterationBucket, fingerprint, itID)
					createChannel(t, acctestIterationBucket, acctestIterationChannel, itID)
				},
				Config: testConfig(testAccPackerIterationAlpineProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePackerIteration_revokedIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.ubuntu"
	fingerprint := "43"
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestIterationUbuntuBucket, acctestIterationChannel, false)
			deleteIteration(t, acctestIterationUbuntuBucket, fingerprint, false)
			deleteBucket(t, acctestIterationUbuntuBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestIterationUbuntuBucket)
					upsertIteration(t, acctestIterationUbuntuBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestIterationUbuntuBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestIterationUbuntuBucket, fingerprint, itID)
					createChannel(t, acctestIterationUbuntuBucket, acctestIterationChannel, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, acctestIterationUbuntuBucket, revokeAt)
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config: testConfig(testAccPackerIterationUbuntuProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "revoke_at", revokeAt.String()),
				),
			},
		},
	})
}
