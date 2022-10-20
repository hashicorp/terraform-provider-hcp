package consul

import (
	"fmt"
	"strings"

	semver "github.com/hashicorp/go-version"
	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
)

// RecommendedVersion returns the recommended version of Consul
func RecommendedVersion(versions []*consulmodels.HashicorpCloudConsul20210204Version) string {
	var defaultVersion string

	for _, v := range versions {
		defaultVersion = v.Version

		if *v.Status == consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED {
			return defaultVersion
		}
	}

	return defaultVersion
}

// IsValidVersion determines that a given version string is contained within the slice of
// available Consul versions.
func IsValidVersion(version string, versions []*consulmodels.HashicorpCloudConsul20210204Version) bool {
	for _, v := range versions {
		if version == v.Version {
			return true
		}
	}

	return false
}

// VersionsToString converts a slice of version pointers to a string of their comma delimited values.
func VersionsToString(versions []*consulmodels.HashicorpCloudConsul20210204Version) string {
	var recommendedVersion string
	var otherVersions []string

	for _, v := range versions {
		if v == nil {
			continue
		}

		if *v.Status == consulmodels.HashicorpCloudConsul20210204VersionStatusRECOMMENDED {
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

// GetLatestPatch parses a list of version strings and returns the latest patch version and a found bool.
//
// E.g. Given the following slice of versions: ["1.11.1", "1.12.2", "1.13.3", "1.14.0"]
// GetLatestPatch("1.13.1", versions) would return ("1.13.3", true)
// GetLatestPatch("1.10.1", versions) would return ("", false)
func GetLatestPatch(version string, versions []*consulmodels.HashicorpCloudConsul20210204Version) (patch string, found bool) {
	// Convert versions to semver and sort them from oldest -> newest.
	// E.g. [1.11.x, 1.12.x, 1.13.x, ...]
	var semvers []*semver.Version
	for _, v := range versions {
		sv, err := semver.NewSemver(v.Version)
		if err != nil {
			// Ignore invalid versions.
			continue
		}

		semvers = append(semvers, sv)
	}

	v, err := semver.NewSemver(version)
	if err != nil {
		return "", false
	}

	for _, sv := range semvers {
		// If the requested version is less than the currently
		// evaluated semver, skip.
		if v.GreaterThan(sv) {
			continue
		}

		// If the requested version's minor does not equal the
		// currently evaluated semver, skip.
		if v.Segments()[1] != sv.Segments()[1] {
			continue
		}

		// Set the patch version to the currently evaluated semver.
		patch = sv.String()
		found = true
	}

	return patch, found
}
