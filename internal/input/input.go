// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input

import (
	"strings"
)

// NormalizeVersion ensures the version starts with a 'v'
func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}
