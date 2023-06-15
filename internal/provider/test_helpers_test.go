// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func testAccCheckFullURL(name, key, port string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		ep := rs.Primary.Attributes[key]

		if !strings.HasPrefix(ep, "https://") {
			return fmt.Errorf("URL missing scheme")
		}

		if port != "" {
			if !strings.HasSuffix(ep, fmt.Sprintf(":%s", port)) {
				return fmt.Errorf("URL missing port")
			}
		}

		return nil
	}
}

// If the resource is not found, the value will be nil and an error is returned.
// If the attribute is not found, the value will be a blank string, but an error will still be returned.
func testAccGetAttributeFromResourceInState(resourceName string, attribute string, state *terraform.State) (*string, error) {
	resources := state.RootModule().Resources

	resource, ok := resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Resource %q not found in the present state", resourceName)
	}

	value, ok := resource.Primary.Attributes[attribute]
	if !ok {
		return &value, fmt.Errorf("Resource %q does not have an attribute named %q in the present state", resourceName, attribute)
	}

	return &value, nil
}

// Returns a best-effort location from the state of a given resource.
// Will return the default location even if the resource isn't found.
func testAccGetLocationFromState(resourceName string, state *terraform.State) (*sharedmodels.HashicorpCloudLocationLocation, error) {

	client := testAccProvider.Meta().(*clients.Client)

	projectIDFromState, _ := testAccGetAttributeFromResourceInState(resourceName, "project_id", state)
	if projectIDFromState == nil {
		return &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      client.Config.ProjectID,
		}, fmt.Errorf("Resource %q not found in the present state", resourceName)
	}

	projectID, _ := GetProjectID(*projectIDFromState, client.Config.OrganizationID)

	return &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}, nil
}

func testAccCreateSlug(testName string) string {
	suffix := fmt.Sprintf("-%s", time.Now().Format("0601021504"))
	return fmt.Sprintf("%.*s%s", 36-len(suffix), testName, suffix)
}

// TODO: Add support for blocks
type testAccConfigBuilderInterface interface {
	IsData() bool
	ResourceType() string
	UniqueName() string
	ResourceName() string
	AttributeRef(string) string
	Attributes() map[string]string
}

func testAccConfigBuildersToString(builders ...testAccConfigBuilderInterface) string {
	config := ""

	for _, cb := range builders {
		rOrD := "resource"
		if cb.IsData() {
			rOrD = "data"
		}

		attributesString := ""
		for key, value := range cb.Attributes() {
			if key != "" && value != "" {
				attributesString += fmt.Sprintf("	%s = %s\n", key, value)
			}
		}

		config += fmt.Sprintf(`
%s %q %q {
%s
}
`, rOrD, cb.ResourceType(), cb.UniqueName(), attributesString)
	}
	return config
}

type testAccConfigBuilder struct {
	isData       bool
	resourceType string
	uniqueName   string
	attributes   map[string]string
	// Attribute values must be as they would be in the config file.
	// Ex: "value" can be represented in Go with `"value"` or fmt.Sprintf("%q", "value")
	// An empty string is equivalent to the attribute not being present in the map.
}

var _ testAccConfigBuilderInterface = testAccConfigBuilder{}

func (b testAccConfigBuilder) IsData() bool {
	return b.isData
}

func (b testAccConfigBuilder) ResourceType() string {
	return b.resourceType
}

func (b testAccConfigBuilder) UniqueName() string {
	return b.uniqueName
}

func (b testAccConfigBuilder) ResourceName() string {
	return fmt.Sprintf("%s.%s", b.ResourceType(), b.UniqueName())
}

func (b testAccConfigBuilder) Attributes() map[string]string {
	return b.attributes
}

func (b testAccConfigBuilder) AttributeRef(path string) string {
	return fmt.Sprintf("%s.%s", b.ResourceName(), path)
}
