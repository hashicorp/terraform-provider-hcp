package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckFullURL(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		ep := rs.Primary.Attributes[key]

		if !strings.HasPrefix(ep, "https://") {
			return fmt.Errorf("URL missing scheme")
		}

		if !strings.HasSuffix(ep, ":8200") {
			return fmt.Errorf("URL missing port")
		}

		return nil
	}
}
