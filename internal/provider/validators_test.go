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
