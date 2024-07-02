// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bucket_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testclient"
)

func TestAccPackerBucketResource(t *testing.T) {
	bucketName := "test-bucket"
	updatedBucketName := "new-test-bucket"
	var createdAt string
	var newCreatedAt string
	// A location is required to upsert the Packer Registry
	// Because of this we have to verify that the acceptance test value is set here, because the acceptance test check normally only occurs inside of resource.Test
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}
	loc := acctest.DefaultProjectLocation(t)
	projectID := loc.GetProjectID()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)
			testclient.UpsertRegistry(t, loc, nil)
		},
		Steps: []resource.TestStep{
			{
				Config: NewPackerBucketResourceConfigBuilder("example").
					WithName(bucketName).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "name", bucketName),
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "resource_name",
						fmt.Sprintf("packer/project/%s/bucket/%s", projectID, bucketName)),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket.example", "created_at"),
					testAccPackerBucketSaveCreatedAt("hcp_packer_bucket.example", &createdAt),
				),
			},
			{
				// Test that bucket can be imported into state
				ResourceName:                         "hcp_packer_bucket.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccPackerBucketImportID,
				ImportStateVerify:                    true,
			},
			{
				Config: NewPackerBucketResourceConfigBuilder("example").
					WithName(updatedBucketName).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "name", updatedBucketName),
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_packer_bucket.example", "resource_name",
						fmt.Sprintf("packer/project/%s/bucket/%s", projectID, updatedBucketName)),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket.example", "created_at"),
					testAccPackerBucketSaveCreatedAt("hcp_packer_bucket.example", &newCreatedAt),
					func(_ *terraform.State) error {
						if newCreatedAt == createdAt {
							return fmt.Errorf("%s %s created_at for both buckets match, indicating resource wasn't recreated", newCreatedAt, createdAt)
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccPackerBucketImportID retrieves the resource_name so that it can be imported.
func testAccPackerBucketImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_packer_bucket.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_name"]
	if !ok {
		return "", fmt.Errorf("resource_name not set")
	}

	return id, nil
}

func testAccPackerBucketSaveCreatedAt(resourceName string, createdAtPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		createdAt := rs.Primary.Attributes["created_at"]

		// assign the response project to the pointer
		*createdAtPtr = createdAt
		return nil
	}
}

type PackerBucketResourceConfigBuilder struct {
	terraformResourceName string
	name                  string
	projectID             string
}

func NewPackerBucketResourceConfigBuilder(terraformResourceName string) PackerBucketResourceConfigBuilder {
	return PackerBucketResourceConfigBuilder{
		terraformResourceName: terraformResourceName,
	}
}
func (b PackerBucketResourceConfigBuilder) WithName(name string) PackerBucketResourceConfigBuilder {
	b.name = name
	return b
}
func (b PackerBucketResourceConfigBuilder) WithProjectID(projectID string) PackerBucketResourceConfigBuilder {
	b.projectID = projectID
	return b
}

func (b PackerBucketResourceConfigBuilder) Build() string {
	projectIDText := ""
	if b.projectID != "" {
		projectIDText = fmt.Sprintf("project id %q", b.projectID)
	}
	config := fmt.Sprintf(`
resource "hcp_packer_bucket" "%s" {
	name = %q
	%s
	
}`,
		b.terraformResourceName,
		b.name,
		projectIDText,
	)
	return config
}
