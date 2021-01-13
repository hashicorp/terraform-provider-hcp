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
