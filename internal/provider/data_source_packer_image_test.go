// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	acctestImageBucket       = fmt.Sprintf("alpine-acc-imagetest-%s", time.Now().Format("200601021504"))
	acctestUbuntuImageBucket = fmt.Sprintf("ubuntu-acc-imagetest-%s", time.Now().Format("200601021504"))
	acctestArchImageBucket   = fmt.Sprintf("arch-acc-imagetest-%s", time.Now().Format("200601021504"))
	acctestImageChannel      = "production-image-test"
	componentType            = "amazon-ebs.example"
)

var (
	testAccPackerImageAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.alpine-imagetest.id
		region         = "us-east-1"
		component_type = %q
	}

	# we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine-imagetest.revoke_at
	}
`, acctestImageBucket, acctestImageChannel, acctestImageBucket, componentType)

	testAccPackerImageAlpineProductionError = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.alpine-imagetest.id
		region         = "us-east-1"
		component_type = "amazon-ebs.do-not-exist"
	}

	# we make sure that this won't fail even when revoke_at is not set
	output "revoke_at" {
  		value = data.hcp_packer_iteration.alpine-imagetest.revoke_at
	}
`, acctestImageBucket, acctestImageChannel, acctestImageBucket)

	testAccPackerImageUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "ubuntu-foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.ubuntu-imagetest.id
		region         = "us-east-1"
	}
`, acctestUbuntuImageBucket, acctestImageChannel, acctestUbuntuImageBucket)

	testAccPackerImageBothChanAndIter = fmt.Sprintf(`
	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		iteration_id = "234567"
		channel = "chanSlug"
		region = "us-east-1"
	}
`, acctestArchImageBucket)

	testAccPackerImageBothChanAndIterRef = fmt.Sprintf(`
	data "hcp_packer_iteration" "arch-imagetest" {
		bucket_name = %q
		channel = %q
	}

	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		iteration_id = data.hcp_packer_iteration.arch-imagetest.id
		channel = %q
		region = "us-east-1"
	}
`, acctestArchImageBucket, acctestImageChannel, acctestArchImageBucket, acctestImageChannel)

	testAccPackerImageArchProduction = fmt.Sprintf(`
	data "hcp_packer_image" "arch-btw" {
		bucket_name = %q
		cloud_provider = "aws"
		channel = %q
		region = "us-east-1"
	}
`, acctestArchImageBucket, acctestImageChannel)
)

func TestAcc_dataSourcePackerImage(t *testing.T) {
	resourceName := "data.hcp_packer_image.foo"
	fingerprint := "44"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestImageBucket, acctestImageChannel, false)
			deleteIteration(t, acctestImageBucket, fingerprint, false)
			deleteBucket(t, acctestImageBucket, false)
			return nil
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			// Testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, acctestImageBucket)
					upsertIteration(t, acctestImageBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestImageBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestImageBucket, fingerprint, itID)
					createChannel(t, acctestImageBucket, acctestImageChannel, itID)
				},
				Config: testAccPackerImageAlpineProduction,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr("data.hcp_packer_image.foo", "labels.test-key", "test-value"),
				),
			},
			// Testing that filtering non-existent image fails properly
			{
				PlanOnly:    true,
				Config:      testAccPackerImageAlpineProductionError,
				ExpectError: regexp.MustCompile("Error: Unable to load image"),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_revokedIteration(t *testing.T) {
	fingerprint := fmt.Sprintf("%d", rand.Int())
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestUbuntuImageBucket, acctestImageChannel, true)
			deleteIteration(t, acctestUbuntuImageBucket, fingerprint, true)
			deleteBucket(t, acctestUbuntuImageBucket, true)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestUbuntuImageBucket)
					upsertIteration(t, acctestUbuntuImageBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestUbuntuImageBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestUbuntuImageBucket, fingerprint, itID)
					createChannel(t, acctestUbuntuImageBucket, acctestImageChannel, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, acctestUbuntuImageBucket, revokeAt)
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config: testConfig(testAccPackerImageUbuntuProduction),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hcp_packer_image.ubuntu-foo", "revoke_at", revokeAt.String()),
					resource.TestCheckResourceAttr("data.hcp_packer_image.ubuntu-foo", "cloud_image_id", "ami-42"),
				),
			},
		},
	})
}

func TestAcc_dataSourcePackerImage_channelAndIterationIDReject(t *testing.T) {
	fingerprint := "rejectIterationAndChannel"
	configs := []string{
		testAccPackerImageBothChanAndIter,
		testAccPackerImageBothChanAndIterRef,
	}

	for _, cfg := range configs {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				// basically just testing that we don't pass validation here
				{
					PlanOnly: true,
					PreConfig: func() {
						deleteChannel(t, acctestArchImageBucket, acctestImageChannel, false)
						deleteIteration(t, acctestArchImageBucket, fingerprint, false)
						deleteBucket(t, acctestArchImageBucket, false)

						upsertRegistry(t)
						upsertBucket(t, acctestArchImageBucket)
						upsertIteration(t, acctestArchImageBucket, fingerprint)
						itID, err := getIterationIDFromFingerPrint(t, acctestArchImageBucket, fingerprint)
						if err != nil {
							t.Fatal(err.Error())
						}
						upsertBuild(t, acctestArchImageBucket, fingerprint, itID)
						createChannel(t, acctestArchImageBucket, acctestImageChannel, itID)
					},
					Config:      testConfig(cfg),
					ExpectError: regexp.MustCompile("Error: Invalid combination of arguments"),
				},
			},
		})
	}
}

func TestAcc_dataSourcePackerImage_channelAccept(t *testing.T) {
	fingerprint := "acceptChannel"
	resourceName := "data.hcp_packer_image.arch-btw"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestArchImageBucket, acctestImageChannel, false)
			deleteIteration(t, acctestArchImageBucket, fingerprint, false)
			deleteBucket(t, acctestArchImageBucket, false)
			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					upsertRegistry(t)
					upsertBucket(t, acctestArchImageBucket)
					upsertIteration(t, acctestArchImageBucket, fingerprint)
					itID, err := getIterationIDFromFingerPrint(t, acctestArchImageBucket, fingerprint)
					if err != nil {
						t.Fatal(err.Error())
					}
					upsertBuild(t, acctestArchImageBucket, fingerprint, itID)
					createChannel(t, acctestArchImageBucket, acctestImageChannel, itID)
				},
				Config: testConfig(testAccPackerImageArchProduction),
				Check: resource.ComposeTestCheckFunc(
					// build_id is only known at runtime
					// and the test works on a reset value,
					// therefore we can only check it's set
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
				),
			},
		},
	})
}
