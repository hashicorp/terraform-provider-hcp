// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parsePeeringResourceID(t *testing.T) {
	defaultProjectID := "e20ad934-b88a-4897-a58e-d8318dd43cc3"
	tests := map[string]struct {
		input             string
		expectedHvnID     string
		expectedPeeringID string
		expectedProjectID string
		hasErr            bool
	}{
		"invalid ID format": {
			input:             "testid",
			expectedHvnID:     "",
			expectedPeeringID: "",
			hasErr:            true,
		},
		"no hvn_id in ID": {
			input:             ":my-peering-id",
			expectedHvnID:     "",
			expectedPeeringID: "",
			hasErr:            true,
		},
		"no peering_id in ID": {
			input:             "my-hvn-id:",
			expectedHvnID:     "",
			expectedPeeringID: "",
			hasErr:            true,
		},
		"valid ID format": {
			input:             "my-hvn-id:my-peering-id",
			expectedHvnID:     "my-hvn-id",
			expectedPeeringID: "my-peering-id",
			expectedProjectID: defaultProjectID,
		},
		"valid ID format with project ID": {
			input:             "ca69d5ff-68c1-4b40-b4fe-b0a1fa80382c:my-hvn-id:my-peering-id",
			expectedHvnID:     "my-hvn-id",
			expectedPeeringID: "my-peering-id",
			expectedProjectID: "ca69d5ff-68c1-4b40-b4fe-b0a1fa80382c",
		},
	}
	for n, tc := range tests {
		t.Run(n, func(*testing.T) {
			r := require.New(t)
			projectID, hvnID, peeringID, err := parsePeeringResourceID(tc.input, defaultProjectID)

			if tc.hasErr {
				r.Error(err)
			} else {
				r.NoError(err)
			}
			r.Equal(tc.expectedHvnID, hvnID)
			r.Equal(tc.expectedPeeringID, peeringID)
			r.Equal(tc.expectedProjectID, projectID)

		})
	}
}
