package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	acctestImageBucket  = "alpine-acctest-imagetest"
	acctestImageChannel = "production-image-test"
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
					createChannel(t, acctestImageBucket, acctestImageChannel)
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
