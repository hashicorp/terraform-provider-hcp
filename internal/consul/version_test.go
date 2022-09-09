package consul

import (
	"testing"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	"github.com/stretchr/testify/require"
)

func Test_RecommendedVersion(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    []*consulmodels.HashicorpCloudConsul20210204Version
	}{
		"with a recommended version": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
			},
			expected: "v1.9.0",
		},
		"without a recommended version": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
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
		validVersions []*consulmodels.HashicorpCloudConsul20210204Version
	}{
		"with a valid version": {
			version: "v1.9.0",
			validVersions: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
			},
			expected: true,
		},
		"with an invalid version": {
			version: "v1.8.0",
			validVersions: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
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

func Test_VersionsToString(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    []*consulmodels.HashicorpCloudConsul20210204Version
	}{
		"with a recommended version": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
			},
			expected: "v1.9.0 (recommended), v1.8.6, v1.8.4",
		},
		"without a recommended version": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.8.6",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
				{
					Version: "v1.8.4",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusAVAILABLE.Pointer(),
				},
			},
			expected: "v1.8.6, v1.8.4",
		},
		"no other versions but recommended": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
			},
			expected: "v1.9.0",
		},
		"nil input": {
			input:    nil,
			expected: "",
		},
		"nil values": {
			input: []*consulmodels.HashicorpCloudConsul20210204Version{
				nil,
				{
					Version: "v1.9.0",
					Status:  consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED.Pointer(),
				},
				nil,
			},
			expected: "v1.9.0",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := VersionsToString(tc.input)
			r.Equal(tc.expected, result)
		})
	}
}
