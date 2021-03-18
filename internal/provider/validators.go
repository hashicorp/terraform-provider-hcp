package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-cty/cty"
	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2021-02-04/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// validateStringNotEmpty ensures a given string is non-empty.
func validateStringNotEmpty(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if v.(string) == "" {
		msg := "cannot be empty"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateStringInSlice returns a func which ensures the string value is a contained in the given slice.
// If ignoreCase is set the strings will be compared as lowercase.
// Adapted from terraform-plugin-sdk validate.StringInSlice
// https://github.com/hashicorp/terraform-plugin-sdk/blob/98ba036fe5895876219331532140d3d8cf239594/helper/validation/strings.go#L132
func validateStringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diagnostics diag.Diagnostics

		value := v.(string)

		for _, validString := range valid {
			if v == validString || (ignoreCase && strings.ToLower(value) == strings.ToLower(validString)) {
				return diagnostics
			}
		}

		msg := fmt.Sprintf("expected %s to be one of %v", value, valid)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
		return diagnostics
	}
}

// validateSemVer ensures a specified string is a SemVer.
func validateSemVer(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !regexp.MustCompile(`^v?\d+.\d+.\d+$`).MatchString(v.(string)) {
		msg := "must be a valid semver"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateSlugID validates that the string value matches the HCP requirements for
// a user-settable slug.
func validateSlugID(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`).MatchString(v.(string)) {
		msg := "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateDatacenter validates that the string value matches the HCP requirements for
// a Consul datacenter name.
func validateDatacenter(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !regexp.MustCompile(`^[-_\da-z]{3,36}$`).MatchString(v.(string)) {
		msg := "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateConsulClusterTier(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := consulmodels.HashicorpCloudConsul20210204ClusterConfigTier(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		expectedEnumList := regexp.MustCompile(`\[.*\]`).FindStringSubmatch(err.Error())
		msg := fmt.Sprintf("expected %v to be one of %v", v, expectedEnumList[0])
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value can be case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateConsulClusterSize(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := consulmodels.HashicorpCloudConsul20210204CapacityConfigSize(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		expectedEnumList := regexp.MustCompile(`\[.*\]`).FindStringSubmatch(err.Error())
		msg := fmt.Sprintf("expected %v to be one of %v", v, expectedEnumList[0])
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value can be case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}
