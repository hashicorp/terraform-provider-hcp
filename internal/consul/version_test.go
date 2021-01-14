package consul

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RecommendedVersion(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    []Version
	}{
		"with a recommended version": {
			input: []Version{
				{
					Version: "v1.9.0",
					Status:  "RECOMMENDED",
				},
				{
					Version: "v1.8.6",
					Status:  "AVAILABLE",
				},
				{
					Version: "v1.8.4",
					Status:  "AVAILABLE",
				},
			},
			expected: "v1.9.0",
		},
		"without a recommended version": {
			input: []Version{
				{
					Version: "v1.9.0",
					Status:  "AVAILABLE",
				},
				{
					Version: "v1.8.6",
					Status:  "AVAILABLE",
				},
				{
					Version: "v1.8.4",
					Status:  "AVAILABLE",
				},
			},
			expected: "v1.8.4",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := RecommendedVersion(tc.input)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_IsValidVersion(t *testing.T) {
	tcs := map[string]struct {
		expected      bool
		version       string
		validVersions []Version
	}{
		"with a valid version": {
			version: "v1.9.0",
			validVersions: []Version{
				{
					Version: "v1.9.0",
					Status:  "RECOMMENDED",
				},
				{
					Version: "v1.8.6",
					Status:  "AVAILABLE",
				},
				{
					Version: "v1.8.4",
					Status:  "AVAILABLE",
				},
			},
			expected: true,
		},
		"with an invalid version": {
			version: "v1.8.0",
			validVersions: []Version{
				{
					Version: "v1.9.0",
					Status:  "RECOMMENDED",
				},
				{
					Version: "v1.8.6",
					Status:  "AVAILABLE",
				},
				{
					Version: "v1.8.4",
					Status:  "AVAILABLE",
				},
			},
			expected: false,
		},
		"with no valid versions": {
			version:       "v1.8.0",
			validVersions: nil,
			expected:      false,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := IsValidVersion(tc.version, tc.validVersions)
			r.Equal(tc.expected, result)
		})
	}
}

func Test_NormalizeVersion(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    string
	}{
		"with a prefixed v": {
			input:    "v1.9.0",
			expected: "v1.9.0",
		},
		"without a prefixed v": {
			input:    "1.9.0",
			expected: "v1.9.0",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := NormalizeVersion(tc.input)
			r.Equal(tc.expected, result)
		})
	}
}
