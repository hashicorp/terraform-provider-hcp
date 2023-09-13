// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"fmt"
	"strings"
	"time"
)

var peeringDefaultTimeout = time.Minute * 1
var peeringCreateTimeout = time.Minute * 35
var peeringDeleteTimeout = time.Minute * 35

func parsePeeringResourceID(resourceID, clientProjectID string) (projectID, hvnID, peeringID string, err error) {
	idParts := strings.SplitN(resourceID, ":", 3)

	if len(idParts) == 3 { // {project_id}:{hvn_id}:{peering_id}
		if idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
			return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected {project_id}:{hvn_id}:{peering_id}", resourceID)
		}
		return idParts[0], idParts[1], idParts[2], nil
	} else if len(idParts) == 2 { // {hvn_id}:{peering_id}
		if idParts[0] == "" || idParts[1] == "" {
			return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{peering_id}", resourceID)
		}
		projectID, err = GetProjectID(projectID, clientProjectID)
		if err != nil {
			return "", "", "", fmt.Errorf("unable to retrieve project ID: %v", err)
		}
		return projectID, idParts[0], idParts[1], nil
	} else {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{peering_id} or {project_id}:{hvn_id}:{peering_id}", resourceID)
	}
}
