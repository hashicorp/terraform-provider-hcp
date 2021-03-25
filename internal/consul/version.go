package consul

import (
	"fmt"
	"strings"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
)

// RecommendedVersion returns the recommended version of Consul
func RecommendedVersion(versions []*consulmodels.HashicorpCloudConsul20200826Version) string {
	var defaultVersion string

	for _, v := range versions {
		defaultVersion = v.Version

		if v.Status == "RECOMMENDED" {
			return defaultVersion
		}
	}

	return defaultVersion
}

// IsValidVersion determines that a given version string is contained within the slice of
// available Consul versions.
func IsValidVersion(version string, versions []*consulmodels.HashicorpCloudConsul20200826Version) bool {
	for _, v := range versions {
		if version == v.Version {
			return true
		}
	}

	return false
}

// VersionsToString converts a slice of version pointers to a string of their comma delimited values.
func VersionsToString(versions []*consulmodels.HashicorpCloudConsul20200826Version) string {
	var recommendedVersion string
	var otherVersions []string

	for _, v := range versions {
		if v == nil {
			continue
		}

		if v.Status == consulmodels.HashicorpCloudConsul20200826VersionStatusRECOMMENDED {
			recommendedVersion = v.Version
		} else {
			otherVersions = append(otherVersions, v.Version)
		}
	}

	// No other versions found, return recommended even if it's empty
	if len(otherVersions) == 0 {
		return recommendedVersion
	}

	// No recommended found, return others
	if recommendedVersion == "" {
		return strings.Join(otherVersions, ", ")
	}

	return fmt.Sprintf("%s (recommended), %s", recommendedVersion, strings.Join(otherVersions, ", "))
}
