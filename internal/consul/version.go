package consul

import (
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

// NormalizeVersion ensures the version starts with a 'v'
func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}
