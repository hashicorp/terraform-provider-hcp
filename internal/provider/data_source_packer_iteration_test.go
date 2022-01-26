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
	acctestIterationBucket       = "alpine-acctest-itertest"
	acctestIterationUbuntuBucket = "ubuntu-acctest-itertest"
	acctestIterationChannel      = "production-iter-test"
)

var (
	testAccPackerIterationAlpineProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "alpine" {
		bucket_name  = %q
		channel = %q
	}`, acctestIterationBucket, acctestIterationChannel)
	testAccPackerIterationUbuntuProduction = fmt.Sprintf(`
	data "hcp_packer_iteration" "ubuntu" {
		bucket_name  = %q
		channel = %q
	}`, acctestIterationUbuntuBucket, acctestIterationChannel)
)

func TestAcc_dataSourcePackerIteration(t *testing.T) {
	resourceName := "data.hcp_packer_iteration.alpine"
	fingerprint := "43"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
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
	fingerprint := fmt.Sprintf("%d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			deleteChannel(t, acctestIterationUbuntuBucket, acctestIterationChannel, false)
			deleteIteration(t, acctestIterationUbuntuBucket, fingerprint, false)
			deleteBucket(t, acctestIterationUbuntuBucket, false)
			return nil
		},

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.
			{
				PreConfig: func() {
					// CheckDestroy doesn't get called when the test fails and doesn't
					// produce any tf state. In this case we destroy any existing resource
					// before creating them.
					deleteChannel(t, acctestIterationUbuntuBucket, acctestIterationChannel, false)
					deleteIteration(t, acctestIterationUbuntuBucket, fingerprint, false)
					deleteBucket(t, acctestIterationUbuntuBucket, false)

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
					revokeIteration(t, itID, acctestIterationUbuntuBucket, "5s")
					// Sleep to make sure the iteration is revoked when we test
					time.Sleep(5 * time.Second)
				},
				Config:      testConfig(testAccPackerIterationUbuntuProduction),
				ExpectError: regexp.MustCompile(`Error: the iteration (\d|\w){26} assigned to channel (\w|\W)* is revoked and can not be used. A valid iteration must be assigned to this channel before proceeding`),
			},
		},
	})
}
