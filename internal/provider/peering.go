package provider

import (
	"fmt"
	"strings"
	"time"
)

var peeringDefaultTimeout = time.Minute * 1
var peeringCreateTimeout = time.Minute * 35
var peeringDeleteTimeout = time.Minute * 35

func parsePeeringResourceID(id string) (hvnID, peeringID string, err error) {
	idParts := strings.SplitN(id, ":", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{peering_id}", id)
	}
	return idParts[0], idParts[1], nil
}
