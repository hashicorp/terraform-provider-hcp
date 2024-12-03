// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// nolint:unused
package providersdkv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

// Checks that the atrribute's value is not the same as diffVal
func testAccCheckResourceAttrDifferent(resourceName string, attribute string, diffVal string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateValPtr, err := testAccGetAttributeFromResourceInState(resourceName, attribute, state)
		if err != nil {
			return err
		}

		if *stateValPtr == diffVal {
			return fmt.Errorf("%s: Attribute '%s' expected to not be %#v, but it was", resourceName, attribute, diffVal)
		}

		return nil
	}
}

// Same as `testAccCheckResourceAttrDifferent`, but diffVal is a pointer that is read at check-time
func testAccCheckResourceAttrPtrDifferent(resourceName string, attribute string, diffVal *string) resource.TestCheckFunc {
	if diffVal == nil {
		panic("diffVal cannot be nil")
	}
	return func(state *terraform.State) error {
		return testAccCheckResourceAttrDifferent(resourceName, attribute, *diffVal)(state)
	}
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

// TODO: Add support for block-type attributes
type testAccConfigBuilderInterface interface {
	// The fully-qualified name of a block that can be referenced, like
	// `hcp_packer_channel.prod` (a resource) or
	// `data.hcp_packer_version.latest` (a data source).
	// Panics if the block cannot be referenced.
	BlockName() string
	// Unquoted Block identifier, like `resource`, `data`, or `output`
	BlockIdentifier() string
	// Unquoted Block labels, like the resource type (`hcp_packer_channel`), or
	// data source type. Includes the UniqueName as the last item if
	// HasUniqueName is true.
	BlockLabels() []string
	// Unquoted unique name for resources/data sources/inputs/outputs.
	// Panics if HasUniqueName is false.
	UniqueName() string

	Attributes() map[string]string
	// Should return `BlockName()+"."+attributeName` for blocks that can be
	// referenced.
	// Panics if the block cannot be referenced.
	AttributeRef(string) string
}

func testAccConfigBuildersToString(builders ...testAccConfigBuilderInterface) string {
	config := ""

	for _, cb := range builders {
		labelsString := ""
		for _, label := range cb.BlockLabels() {
			labelsString += fmt.Sprintf("%q ", label)
		}

		attributesString := ""
		for key, value := range cb.Attributes() {
			if key != "" && value != "" {
				attributesString += fmt.Sprintf("	%s = %s\n", key, value)
			}
		}

		config += fmt.Sprintf(`
%s %s{
%s
}
`,
			cb.BlockIdentifier(), labelsString,
			attributesString,
		)
	}
	return config
}

// Generic ConfigBuilder for Blocks that have a UniqueName and no other labels
// like `output` and `variable`
type testAccGenericNamedBlockConfigBuilder struct {
	// Set to true if the block type can be referenced
	canReference bool
	// If canReference is true, set this value to the reference identifier.
	// Ex: `var` is the reference identifier for the `variable` block type.
	referenceIdentifier string
	blockIdentifier     string
	uniqueName          string
	attributes          map[string]string
}

var _ testAccConfigBuilderInterface = testAccGenericNamedBlockConfigBuilder{}

func (b testAccGenericNamedBlockConfigBuilder) BlockName() string {
	if !b.canReference {
		panic(fmt.Errorf("`%s` blocks cannot be referenced", b.BlockIdentifier()))
	}

	return fmt.Sprintf("%s.%s", b.referenceIdentifier, b.uniqueName)
}

func (b testAccGenericNamedBlockConfigBuilder) BlockIdentifier() string {
	return b.blockIdentifier
}

func (b testAccGenericNamedBlockConfigBuilder) BlockLabels() []string {
	return []string{b.uniqueName}
}

func (b testAccGenericNamedBlockConfigBuilder) UniqueName() string {
	return b.uniqueName
}

func (b testAccGenericNamedBlockConfigBuilder) HasUniqueName() bool {
	return true
}

func (b testAccGenericNamedBlockConfigBuilder) Attributes() map[string]string {
	return b.attributes
}

func (b testAccGenericNamedBlockConfigBuilder) AttributeRef(attributeName string) string {
	if !b.canReference {
		panic(fmt.Errorf("`%s` blocks cannot be referenced", b.BlockIdentifier()))
	}

	return fmt.Sprintf("%s.%s", b.BlockName(), attributeName)
}

type testAccResourceConfigBuilder struct {
	isData       bool
	resourceType string
	uniqueName   string
	attributes   map[string]string
	// Attribute values must be as they would be in the config file.
	// Ex: "value" can be represented in Go with `"value"` or fmt.Sprintf("%q", "value")
	// An empty string is equivalent to the attribute not being present in the map.
}

var _ testAccConfigBuilderInterface = testAccResourceConfigBuilder{}

func (b testAccResourceConfigBuilder) BlockName() string {
	if b.isData {
		return fmt.Sprintf("data.%s.%s", b.resourceType, b.UniqueName())
	}

	return fmt.Sprintf("%s.%s", b.resourceType, b.UniqueName())
}

func (b testAccResourceConfigBuilder) BlockIdentifier() string {
	if b.isData {
		return "data"
	}

	return "resource"
}

func (b testAccResourceConfigBuilder) BlockLabels() []string {
	return []string{b.resourceType, b.UniqueName()}
}

func (b testAccResourceConfigBuilder) UniqueName() string {
	return b.uniqueName
}

func (b testAccResourceConfigBuilder) HasUniqueName() bool {
	return true
}

func (b testAccResourceConfigBuilder) Attributes() map[string]string {
	return b.attributes
}

func (b testAccResourceConfigBuilder) AttributeRef(path string) string {
	return fmt.Sprintf("%s.%s", b.BlockName(), path)
}

// Used to create a dummy non-empty state so that `CheckDestroy` can be used to
// clean up resources created in `PreCheck` for tests that don't generate a
// non-empty state on their own.
const testAccConfigDummyNonemptyState string = `resource "dummy_state" "dummy" {}`

// Contains a dummy resource that is used in `testAccConfigDummyNonemptyState`
func testAccNewDummyProvider() *schema.Provider {
	setDummyID := func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
		d.SetId("dummy_id")
		return nil
	}
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"dummy_state": {
				CreateContext: setDummyID,
				ReadContext:   setDummyID,
				DeleteContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
			},
		},
	}
}
