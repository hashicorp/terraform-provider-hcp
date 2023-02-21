// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func Test_validateStringNotEmpty(t *testing.T) {
	tcs := map[string]struct {
		expected diag.Diagnostics
		input    string
	}{
		"valid string": {
			input:    "hello",
			expected: nil,
		},
		"empty string": {
			input: "",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "cannot be empty",
					Detail:        "cannot be empty",
					AttributePath: nil,
				},
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := validateStringNotEmpty(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateStringInSlice(t *testing.T) {
	tcs := map[string]struct {
		expected    diag.Diagnostics
		input       string
		ignoreCase  bool
		validValues []string
	}{
		"contains the input (matches case)": {
			input:       "hello",
			expected:    nil,
			ignoreCase:  false,
			validValues: []string{"hello", "bonjour"},
		},
		"contains the input (case invariant)": {
			input:       "HELLO",
			expected:    nil,
			ignoreCase:  true,
			validValues: []string{"hello", "bonjour"},
		},
		"does not contain the input": {
			input: "hello",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("expected hello to be one of %v", []string{"goodbye", "bonjour"}),
					Detail:        fmt.Sprintf("expected hello to be one of %v", []string{"goodbye", "bonjour"}),
					AttributePath: nil,
				},
			},
			ignoreCase:  false,
			validValues: []string{"goodbye", "bonjour"},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := validateStringInSlice(tc.validValues, tc.ignoreCase)(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateSemVer(t *testing.T) {
	tcs := map[string]struct {
		expected diag.Diagnostics
		input    string
	}{
		"valid semver with prefixed v": {
			input:    "v1.2.3",
			expected: nil,
		},
		"valid semver without prefixed v": {
			input:    "1.2.3",
			expected: nil,
		},
		"invalid semver": {
			input: "v1.2.3.4.5",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be a valid semver",
					Detail:        "must be a valid semver",
					AttributePath: nil,
				},
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := validateSemVer(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateSlugID(t *testing.T) {
	tcs := map[string]struct {
		expected diag.Diagnostics
		input    string
	}{
		"valid id": {
			input:    "hello-123",
			expected: nil,
		},
		"empty string": {
			input: "",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					Detail:        "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					AttributePath: nil,
				},
			},
		},
		"invalid characters": {
			input: "test@123",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					Detail:        "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					AttributePath: nil,
				},
			},
		},
		"too short": {
			input: "ab",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					Detail:        "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					AttributePath: nil,
				},
			},
		},
		"too long": {
			input: "abcdefghi1abcdefghi1abcdefghi12345678",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					Detail:        "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					AttributePath: nil,
				},
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := validateSlugID(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateDatacenter(t *testing.T) {
	tcs := map[string]struct {
		expected diag.Diagnostics
		input    string
	}{
		"valid datacenter": {
			input:    "hello-123_456",
			expected: nil,
		},
		"empty string": {
			input: "",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					Detail:        "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					AttributePath: nil,
				},
			},
		},
		"invalid characters": {
			input: "test@123",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					Detail:        "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					AttributePath: nil,
				},
			},
		},
		"uppercase characters": {
			input: "Test123",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					Detail:        "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					AttributePath: nil,
				},
			},
		},
		"too short": {
			input: "ab",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					Detail:        "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					AttributePath: nil,
				},
			},
		},
		"too long": {
			input: "abcdefghi1abcdefghi1abcdefghi12345678",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					Detail:        "must be between 3 and 36 characters in length and contains only lowercase letters, numbers, hyphens, or underscores",
					AttributePath: nil,
				},
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := validateDatacenter(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateConsulClusterTier(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"development": {
			input:    "development",
			expected: nil,
		},
		"standard": {
			input:    "standard",
			expected: nil,
		},
		"plus": {
			input:    "plus",
			expected: nil,
		},
		"valid tier lowercase": {
			input:    "development",
			expected: nil,
		},
		"valid tier uppercase": {
			input:    "STANDARD",
			expected: nil,
		},
		"valid tier mixedcase": {
			input:    "DEVelopment",
			expected: nil,
		},
		"invalid tier": {
			input: "dev",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "expected dev to be one of [DEVELOPMENT STANDARD PLUS PREMIUM]",
					Detail:        "expected dev to be one of [DEVELOPMENT STANDARD PLUS PREMIUM] (value is case-insensitive).",
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateConsulClusterTier(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateConsulClusterSize(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"valid size": {
			input:    "x_small",
			expected: nil,
		},
		"valid size lowercase": {
			input:    "small",
			expected: nil,
		},
		"valid size uppercase": {
			input:    "MEDIUM",
			expected: nil,
		},
		"valid size mixedcase": {
			input:    "LARge",
			expected: nil,
		},
		"invalid tier": {
			input: "med",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "expected med to be one of [X_SMALL SMALL MEDIUM LARGE]",
					Detail:        "expected med to be one of [X_SMALL SMALL MEDIUM LARGE] (value is case-insensitive).",
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateConsulClusterSize(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateConsulClusterCIDR(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"valid IP address": {
			input:    "172.25.16.0/24",
			expected: nil,
		},
		"invalid ip address": {
			input: "invalid",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid address (invalid) of ip_allowlist",
					Detail:        "invalid address (invalid) of ip_allowlist (must be a valid IPV4 CIDR).",
					AttributePath: nil,
				},
			},
		},
		"IPV6 unsupported": {
			input: "2002::1234:abcd:ffff:c0a8:101/64",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid address (2002::1234:abcd:ffff:c0a8:101/64) of ip_allowlist",
					Detail:        "invalid address (2002::1234:abcd:ffff:c0a8:101/64) of ip_allowlist (must be a valid IPV4 CIDR).",
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateConsulClusterCIDR(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateConsulClusterCIDRDescription(t *testing.T) {
	invalidInput := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"valid description": {
			input:    "IPV4 address",
			expected: nil,
		},
		"invalid ip address": {
			input: invalidInput,
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("invalid description (%s) of ip_allowlist", invalidInput),
					Detail:        fmt.Sprintf("invalid description (%s) of ip_allowlist (must be within 255 char).", invalidInput),
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateConsulClusterCIDRDescription(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateVaultClusterTier(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"valid tier lowercase": {
			input:    "dev",
			expected: nil,
		},
		"valid tier uppercase": {
			input:    "STANDARD_SMALL",
			expected: nil,
		},
		"valid tier mixedcase": {
			input:    "StanDard_LargE",
			expected: nil,
		},
		"invalid tier": {
			input: "development",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "expected 'development' to be one of: [dev standard_small standard_medium standard_large starter_small plus_small plus_medium plus_large]",
					Detail:        "expected 'development' to be one of: [dev standard_small standard_medium standard_large starter_small plus_small plus_medium plus_large] (value is case-insensitive).",
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateVaultClusterTier(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateVaultPathsFilter(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"valid path": {
			input:    "valid/path",
			expected: nil,
		},
		"different valid path": {
			input:    "_valid-path/2/2/2/valid",
			expected: nil,
		},
		"invalid path with :": {
			input: "valid/path:",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "paths filter path 'valid/path:' is invalid",
					Detail:        "paths filter path 'valid/path:' is invalid (paths must match regex '\\A[\\w-]+(/[\\w-]+)*\\z').",
					AttributePath: nil,
				},
			},
		},
		"invalid path with trailing /": {
			input: "trailing/",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "paths filter path 'trailing/' is invalid",
					Detail:        "paths filter path 'trailing/' is invalid (paths must match regex '\\A[\\w-]+(/[\\w-]+)*\\z').",
					AttributePath: nil,
				},
			},
		},
		"invalid path with leading /": {
			input: "/leading",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "paths filter path '/leading' is invalid",
					Detail:        "paths filter path '/leading' is invalid (paths must match regex '\\A[\\w-]+(/[\\w-]+)*\\z').",
					AttributePath: nil,
				},
			},
		},
		"invalid empty path": {
			input: "",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "paths filter path '' is invalid",
					Detail:        "paths filter path '' is invalid (paths must match regex '\\A[\\w-]+(/[\\w-]+)*\\z').",
					AttributePath: nil,
				},
			},
		},
	}
	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateVaultPathsFilter(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_validateCIDRBlock(t *testing.T) {
	tcs := map[string]struct {
		input    string
		expected diag.Diagnostics
	}{
		"blank input string": {
			input: "",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"invalid cidr notation": {
			input: "192.168.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"invalid characters": {
			input: "someString",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"valid cidr block": {
			input:    "10.0.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 2": {
			input:    "10.255.255.0/24",
			expected: diag.Diagnostics(nil),
		},
		"invalid cidr notation 2": {
			input: "10.256.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"invalid cidr notation 3": {
			input: "10.0.256.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"invalid characters 2": {
			input: "10.255.255asdfasdfaqsd.250",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"valid cidr block 3": {
			input:    "192.168.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 4": {
			input:    "192.168.255.0/24",
			expected: diag.Diagnostics(nil),
		},
		"invalid cidr notation 4": {
			input: "192.168.256.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "unable to parse string as CIDR notation IP address",
					Detail:        "unable to parse string as CIDR notation IP address",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern": {
			input: "192.0.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"valid cidr block 5": {
			input:    "172.16.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 6": {
			input:    "172.17.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 7": {
			input:    "172.18.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 8": {
			input:    "172.30.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 9": {
			input:    "172.20.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"valid cidr block 10": {
			input:    "172.31.0.0/24",
			expected: diag.Diagnostics(nil),
		},
		"invalid pattern 2": {
			input: "172.15.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern 3": {
			input: "172.32.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern 4": {
			input: "172.192.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern 5": {
			input: "172.255.0.0/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern 6": {
			input: "10.0.0.0/7",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern 7": {
			input: "192.168.0.0/15",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
			},
		},
		"invalid pattern and invalid range": {
			input: "87.70.141.1/22",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					Detail:        "must match pattern of 10.*.*.* with prefix greater than /8,or 172.[16-31].*.* with prefix greater than /12, or 192.168.*.* with prefix greater than /16; where * is any number from [0-255]",
					AttributePath: nil,
				},
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid CIDR range start 87.70.141.1, should have been 87.70.140.0",
					Detail:        "invalid CIDR range start 87.70.141.1, should have been 87.70.140.0",
					AttributePath: nil,
				},
			},
		},
		"invalid range": {
			input: "192.168.255.255/24",
			expected: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid CIDR range start 192.168.255.255, should have been 192.168.255.0",
					Detail:        "invalid CIDR range start 192.168.255.255, should have been 192.168.255.0",
					AttributePath: nil,
				},
			},
		},
		"valid cidr block 11": {
			input:    "172.25.16.0/24",
			expected: diag.Diagnostics(nil),
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)
			result := validateCIDRBlock(tc.input, nil)
			r.Equal(tc.expected, result)
		})
	}
}
