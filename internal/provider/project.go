// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
)

func GetProjectID(resourceProjID, clientProjID string) (string, error) {
	if resourceProjID != "" {
		return resourceProjID, nil
	} else {
		if clientProjID != "" {
			return clientProjID, nil
		} else {
			return "", fmt.Errorf("project ID not defined. Verify that project ID is set either in the provider or in the resource config")
		}
	}
}
