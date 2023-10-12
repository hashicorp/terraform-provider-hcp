package acctest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// CheckDataSourceStateMatchesResourceStateWithIgnores returns a
// resource.TestCheckFunc that compares the state of a data source to that of a
// resource. For example, an acceptance test can use the following config:
//
//	resource "hcp_project" "project" {
//	  name        = %q
//	  description = %q
//	}
//
//	data "hcp_project" "project" {
//	  project = hcp_project.project.resource_id
//	}
//
// Then the following check can be defined:
//
//	  CheckDataSourceStateMatchesResourceStateWithIgnores(
//		 "data.hcp_project.project",
//	     "hcp_project.project",
//	      map[string]struct{}{"project": {}}),
//	  )
func CheckDataSourceStateMatchesResourceStateWithIgnores(dataSourceName, resourceName string, ignoreFields map[string]struct{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", dataSourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		dsAttr := ds.Primary.Attributes
		rsAttr := rs.Primary.Attributes

		errMsg := ""
		// Data sources are often derived from resources, so iterate over the resource fields to
		// make sure all fields are accounted for in the data source.
		// If a field exists in the data source but not in the resource, its expected value should
		// be checked separately.
		for k := range rsAttr {
			if _, ok := ignoreFields[k]; ok {
				continue
			}
			if _, ok := ignoreFields["labels.%"]; ok && strings.HasPrefix(k, "labels.") {
				continue
			}
			if _, ok := ignoreFields["terraform_labels.%"]; ok && strings.HasPrefix(k, "terraform_labels.") {
				continue
			}
			if k == "%" {
				continue
			}
			if dsAttr[k] != rsAttr[k] {
				// ignore data sources where an empty list is being compared against a null list.
				if k[len(k)-1:] == "#" && (dsAttr[k] == "" || dsAttr[k] == "0") && (rsAttr[k] == "" || rsAttr[k] == "0") {
					continue
				}
				errMsg += fmt.Sprintf("%s is %s; want %s\n", k, dsAttr[k], rsAttr[k])
			}
		}

		if errMsg != "" {
			return errors.New(errMsg)
		}

		return nil
	}
}
