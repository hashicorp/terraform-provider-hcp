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
	target, err := semver.NewSemver(version)
	if err != nil {
		return "", false
	}

	// Keep track of the current patch version.
	var currentPatch int

	for _, current := range versions {
		sv, err := semver.NewSemver(current.Version)
		if err != nil {
			// Ignore invalid versions.
			continue
		}

		// If the target version is greater than the currently
		// evaluated semver, skip.
		if target.GreaterThan(sv) {
			continue
		}

		// If the requested version's minor does not equal the
		// currently evaluated semver, skip.
		if target.Segments()[1] != sv.Segments()[1] {
			continue
		}

		// Check the current patch is greater than the previous patch version.
		p := sv.Segments()[2]
		if p >= currentPatch {
			currentPatch = p
			// Set the patch version to the currently evaluated semver.
			patch = sv.String()
			found = true
		}

	}

	return patch, found
}
