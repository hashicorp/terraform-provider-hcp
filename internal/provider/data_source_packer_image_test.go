package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	acctestImageBucket       = "alpine-acctest-imagetest"
	acctestImageUbuntuBucket = "ubuntu-acctest-imagetest"
	acctestImageChannel      = "production-image-test"
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
	}`, acctestImageBucket, acctestImageChannel, acctestImageBucket)
	testAccPackerImageUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu-imagetest" {
		bucket_name  = %q
		channel = %q
	}

	data "hcp_packer_image" "foo" {
		bucket_name    = %q
		cloud_provider = "aws"
		iteration_id   = data.hcp_packer_iteration.ubuntu-imagetest.id
		region         = "us-east-1"
	}`, acctestImageUbuntuBucket, acctestImageChannel, acctestImageUbuntuBucket)
)

func TestAcc_dataSourcePackerImage(t *testing.T) {
	resourceName := "data.hcp_packer_image.foo"
	fingerprint := "44"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			itID := getIterationIDFromFingerPrint(t, acctestImageBucket, fingerprint)
			deleteChannel(t, acctestImageBucket, acctestImageChannel)
			deleteIteration(t, acctestImageBucket, itID)
			deleteBucket(t, acctestImageBucket)
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
					itID := getIterationIDFromFingerPrint(t, acctestImageBucket, fingerprint)
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
		},
	})
}

func TestAcc_dataSourcePackerImage_revokedIteration(t *testing.T) {
	fingerprint := fmt.Sprintf("%d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			itID := getIterationIDFromFingerPrint(t, acctestImageUbuntuBucket, fingerprint)
			deleteChannel(t, acctestImageUbuntuBucket, acctestImageChannel)
			deleteIteration(t, acctestImageUbuntuBucket, itID)
			deleteBucket(t, acctestImageUbuntuBucket)
			return nil
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			// Testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, acctestImageUbuntuBucket)
					upsertIteration(t, acctestImageUbuntuBucket, fingerprint)
					itID := getIterationIDFromFingerPrint(t, acctestImageUbuntuBucket, fingerprint)
					upsertBuild(t, acctestImageUbuntuBucket, fingerprint, itID)
					createChannel(t, acctestImageUbuntuBucket, acctestImageChannel, itID)
					// Schedule revocation to the future, otherwise we won't be able to revoke an iteration that
					// it's assigned to a channel
					revokeIteration(t, itID, acctestImageUbuntuBucket, "5s")
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config:      testAccPackerImageUbuntuProduction,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Error: the iteration (\d|\w){26} is revoked and can not be used`),
			},
		},
	})
}
