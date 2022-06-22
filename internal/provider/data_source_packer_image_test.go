package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	acctestImageBucket       = "alpine-acctest-imagetest"
	acctestUbuntuImageBucket = "ubuntu-acctest-imagetest"
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
		},
	})
}

func TestAcc_dataSourcePackerImage_revokedIteration(t *testing.T) {
	fingerprint := fmt.Sprintf("%d", rand.Int())
	revokeAt := strfmt.DateTime(time.Now().UTC().Add(5 * time.Minute))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// testing that getting a revoked iteration fails properly
			{
				PlanOnly: true,
				PreConfig: func() {
					// CheckDestroy doesn't get called when the test fails and doesn't
					// produce any tf state. In this case we destroy any existing resource
					// before creating them.
					deleteChannel(t, acctestUbuntuImageBucket, acctestImageChannel, false)
					deleteIteration(t, acctestUbuntuImageBucket, fingerprint, false)
					deleteBucket(t, acctestUbuntuImageBucket, false)

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
					resource.TestCheckResourceAttr("data.hcp_packer_image.ubuntu-foo", "cloud_image_id", "error_revoked"),
				),
			},
		},
	})
}
