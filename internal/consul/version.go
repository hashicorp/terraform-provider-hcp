package consul

import "strings"

// NormalizeVersion ensures the version starts with a 'v'
func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}
