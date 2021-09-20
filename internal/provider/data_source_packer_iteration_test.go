package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	acctestIterationBucket  = "alpine-acctest-itertest"
	acctestIterationChannel = "production-iter-test"
)

var (
	testAccPackerIterationAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}`, acctestIterationBucket, acctestIterationChannel)
)

func TestAcc_dataSourcePackerIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.alpine"
	fingerprint := "43"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			itID := getIterationIDFromFingerPrint(t, acctestIterationBucket, fingerprint)
			// delete iteration before channel to ensure hard delete of channel.
			deleteIteration(t, acctestIterationBucket, itID)
			deleteChannel(t, acctestIterationBucket, acctestIterationChannel)
			deleteBucket(t, acctestIterationBucket)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					upsertBucket(t, acctestIterationBucket)
					upsertIteration(t, acctestIterationBucket, fingerprint)
					itID := getIterationIDFromFingerPrint(t, acctestIterationBucket, fingerprint)
					upsertBuild(t, acctestIterationBucket, fingerprint, itID)
					createChannel(t, acctestIterationBucket, acctestIterationChannel)
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
