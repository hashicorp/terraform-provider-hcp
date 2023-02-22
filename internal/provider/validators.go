// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-cty/cty"
	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
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
			if v == validString || (ignoreCase && strings.EqualFold(value, validString)) {
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

// matchesID matches /project/11eabb9f-d2ee-9c80-9483-0242ac110013/hashicorp.consul.cluster/example
func matchesID(id string) bool {
	return regexp.MustCompile(`/project/\b[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\b/hashicorp\.consul\.cluster/.*`).MatchString(id)
}

func matchesSlugID(slugID string) bool {
	return regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`).MatchString(slugID)
}

// validateSlugIDOrID validates that the string value matches the HCP requirements for
// an id or slug id.
func validateSlugIDOrID(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !matchesID(v.(string)) && !matchesSlugID(v.(string)) {
		msg := "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens OR must match /project/uuid/hashicorp.consul.cluster/id format"
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

	if !matchesSlugID(v.(string)) {
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

	// TODO: Update the validation once SDK provides a way to get all valid values for the enum.
	err := consulmodels.HashicorpCloudConsul20210204ClusterConfigTier(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.Replace(enumList, "UNSET ", "", 1)
		msg := fmt.Sprintf("expected %v to be one of %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateConsulClusterSize(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	// TODO: Update the validation once SDK provides a way to get all valid values for the enum.
	err := consulmodels.HashicorpCloudConsul20210204CapacityConfigSize(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.Replace(enumList, "UNSET ", "", 1)
		msg := fmt.Sprintf("expected %v to be one of %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateConsulClusterCIDR(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	addr := v.(string)
	ip, err := netip.ParsePrefix(addr)
	isIPV4 := ip.Addr().Is4()

	if err != nil || !ip.IsValid() || !isIPV4 {
		msg := fmt.Sprintf("invalid address (%v) of ip_allowlist", v)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (must be a valid IPV4 CIDR).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateConsulClusterCIDRDescription(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	description := v.(string)
	if len(description) > 255 {
		msg := fmt.Sprintf("invalid description (%v) of ip_allowlist", v)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (must be within 255 char).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateVaultClusterTier(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := vaultmodels.HashicorpCloudVault20201125Tier(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.ToLower(enumList)
		msg := fmt.Sprintf("expected '%v' to be one of: %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateVaultUpgradeType(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigUpgradeType(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.ToLower(enumList)
		msg := fmt.Sprintf("expected '%v' to be one of: %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateVaultUpgradeWindowDay(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindowDayOfWeek(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.ToLower(enumList)
		msg := fmt.Sprintf("expected '%v' to be one of: %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateVaultUpgradeWindowTime(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	err := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindowTimeWindowUTC(strings.ToUpper(v.(string))).Validate(strfmt.Default)
	if err != nil {
		enumList := regexp.MustCompile(`\[.*\]`).FindString(err.Error())
		expectedEnumList := strings.ToLower(enumList)
		msg := fmt.Sprintf("expected '%v' to be one of: %v", v, expectedEnumList)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + " (value is case-insensitive).",
			AttributePath: path,
		})
	}

	return diagnostics
}

func validateVaultPathsFilter(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	p := v.(string)
	pathRegex := regexp.MustCompile(`\A[\w-]+(/[\w-]+)*\z`)
	if !pathRegex.MatchString(p) {
		msg := fmt.Sprintf("paths filter path '%v' is invalid", p)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg + fmt.Sprintf(" (paths must match regex '%s').", pathRegex.String()),
			AttributePath: path,
		})
	}
	return diagnostics
}

func validateCIDRBlock(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	// validRanges contains the set of IP ranges considered valid.
	var validRanges = []net.IPNet{
		{
			// 10.*.*.*
			IP:   net.IPv4(10, 0, 0, 0),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
		{
			// 192.168.*.*
			IP:   net.IPv4(192, 168, 0, 0),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		},
		{
			// 172.[16-31].*.*
			IP:   net.IPv4(172, 16, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	// parse the string as CIDR notation IP address and prefix length.
	ip, net, err := net.ParseCIDR(v.(string))
	if err != nil {
		msg := "unable to parse string as CIDR notation IP address"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})

		return diagnostics
	}

	// validate if the IP address is contained in one of the expected ranges.
	valid := false
	for _, validRange := range validRanges {
		valueSize, _ := net.Mask.Size()
		validRangeSize, _ := validRange.Mask.Size()
		if validRange.Contains(ip) && valueSize >= validRangeSize {
			// Flip flag if IP is found within any 1 of 3 ranges.
			valid = true
		}
	}

	// Check flag and return an error if the IP address is not contained within
	// any of the expected ranges.
	if !valid {
		msg := fmt.Sprintf("must match pattern of 10.*.*.* with prefix greater than /8," +
			"or 172.[16-31].*.* with prefix greater than /12, or " +
			"192.168.*.* with prefix greater than /16; where * is any number from [0-255]")
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	// Validate the address passed is the start of the CIDR range.
	// This happens after we verify the IP address is a valid RFC 1819
	// range to avoid causing confusion with a misguiding error message.
	if !ip.Equal(net.IP) {
		msg := fmt.Sprintf("invalid CIDR range start %s, should have been %s", ip, net.IP)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// Validate the provided initial admin username for a boundary cluster
func validateBoundaryUsername(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !regexp.MustCompile("^[a-z0-9.-]{3,}$").MatchString(v.(string)) {
		msg := "invalid boundary username; login name must be all-lowercase alphanumeric, period or hyphen, and at least 3 characters."
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// Validate the password for the initial boundary cluster admin user
func validateBoundaryPassword(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if len(v.(string)) < 8 {
		msg := "invalid boundary password; password must be at least 8 characters."
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}
