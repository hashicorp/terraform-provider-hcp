// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"fmt"
	"regexp"
)

var (
	// GroupResourceName is a regex that matches a group resource name
	GroupResourceName = regexp.MustCompile(`^iam/organization/.+/group/.+$`)
)

func ResourceName(groupName, orgID string) string {
	rn := groupName
	if !GroupResourceName.MatchString(rn) {
		rn = fmt.Sprintf("iam/organization/%s/group/%s", orgID, groupName)
	}

	return rn
}
