// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package helpers

var DisabledTiers = []string{"STARTER_SMALL"}

func IsDisabledTier(v string) bool {
	for _, tier := range DisabledTiers {
		if tier == v {
			return true
		}
	}
	return false
}
