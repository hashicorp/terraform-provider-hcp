package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	acctestIterationBucket       = "alpine-acctest-itertest"
	acctestIterationUbuntuBucket = "ubuntu-acctest-itertest"
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
