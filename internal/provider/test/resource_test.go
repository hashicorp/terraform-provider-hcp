package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

// TestAccMultipleProjectsGroupsIamBindings tests the creation of multiple projects, groups, and IAM bindings.
type testMultipleProjectsGroupsIAMInput struct {
	Projects []projectInput
	Groups   []groupInput
	Bindings []iamBindingInput
}

// projectInput defines the structure for project input data.
type projectInput struct {
	Name        string
	Description string
}

// groupInput defines the structure for group input data.
type groupInput struct {
	Name        string
	Description string
}

// iamBindingInput defines the structure for IAM binding input data.
type iamBindingInput struct {
	Role        string
	ProjectName string
	GroupName   string
}

// setTestAccMultipleProjectsGroupsIAM generates a Terraform configuration string for multiple projects, groups, and IAM bindings.
func setTestAccMultipleProjectsGroupsIAM(input *testMultipleProjectsGroupsIAMInput) (string, error) {
	var buf strings.Builder

	for i := 0; i < len(input.Projects); i++ {
		proj := input.Projects[i]
		group := input.Groups[i]
		binding := input.Bindings[i]

		// Write a project
		buf.WriteString(fmt.Sprintf(`
			resource "hcp_project" "p%d" {
  				name = %q
  				description = %q
			}
		`, i, proj.Name, proj.Description),
		)

		// Write a group
		buf.WriteString(fmt.Sprintf(`
			resource "hcp_group" "g%d" {
  				display_name = %q
  				description  = %q
			}
			`, i, group.Name, group.Description),
		)

		// Write a binding
		buf.WriteString(fmt.Sprintf(`
			resource "hcp_project_iam_binding" "b%d" {
  				project_id   = hcp_project.p%d.resource_id
				principal_id = hcp_group.g%d.resource_id
  				role         = %q
			}
			`, i, i, i, binding.Role),
		)
	}

	return buf.String(), nil
}

// TestAccMultipleProjectsGroupsIamBindings tests the creation of multiple projects, groups, and IAM bindings.
func TestAccMultipleProjectsGroupsIamBindings(t *testing.T) {
	t.Parallel()

	n := 1
	input := &testMultipleProjectsGroupsIAMInput{
		Projects: make([]projectInput, n),
		Groups:   make([]groupInput, n),
		Bindings: make([]iamBindingInput, n),
	}

	for i := 0; i < n; i++ {
		projectName := fmt.Sprintf("proj-%s", acctest.RandString(8))
		groupName := fmt.Sprintf("group-%s", acctest.RandString(8))

		input.Projects[i] = projectInput{Name: projectName, Description: acctest.RandString(16)}
		input.Groups[i] = groupInput{Name: groupName, Description: acctest.RandString(16)}
		input.Bindings[i] = iamBindingInput{
			Role:        "roles/viewer",
			ProjectName: projectName,
			GroupName:   groupName,
		}
	}

	config, err := setTestAccMultipleProjectsGroupsIAM(input)
	if err != nil {
		t.Fatalf("config generation failed: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testConfig(config),
				Check: resource.ComposeTestCheckFunc(func(state *terraform.State) error {
					for i := 0; i < len(input.Projects); i++ {
						// Check if project exists
						projectName := fmt.Sprintf("hcp_project.p%d", i)
						if _, ok := state.RootModule().Resources[projectName]; !ok {
							return fmt.Errorf("project resource %s not found", projectName)
						}

						// Check if group exists
						groupName := fmt.Sprintf("hcp_group.g%d", i)
						if _, ok := state.RootModule().Resources[groupName]; !ok {
							return fmt.Errorf("group resource %s not found", groupName)
						}

						// Check if IAM binding exists
						bindingName := fmt.Sprintf("hcp_project_iam_binding.b%d", i)
						if _, ok := state.RootModule().Resources[bindingName]; !ok {
							return fmt.Errorf("binding resource %s not found", bindingName)
						}
					}
					return nil
				}),
			},
		},
	})

}

// testConfig is a helper function to trim whitespace from the configuration string.
func testConfig(cfg string) string {
	return strings.TrimSpace(cfg)
}
