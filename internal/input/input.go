// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input

import (
	"regexp"
	"strings"
)

// NormalizeVersion ensures the version starts with a 'v'
func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}

// IsSlug checks if a string is a valid HCP slug, which is 3-36 characters long, begins with a
// letter or number, ends with a letter or number, and contains only [A-Za-z0-9-].
func IsSlug(slug string) bool {
	return regexp.MustCompile(`^[\da-zA-Z][-a-zA-Z\d]{1,34}[\da-zA-Z]$`).MatchString(slug)
}
