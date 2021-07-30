package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parsePeeringResourceID(t *testing.T) {
	tests := map[string]struct {
		input             string
		expectedHvnID     string
		expectedPeeringID string
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
		},
	}
	for n, tc := range tests {
		t.Run(n, func(*testing.T) {
			r := require.New(t)
			hvnID, peeringID, err := parsePeeringResourceID(tc.input)

			if tc.hasErr {
				r.Error(err)
			} else {
				r.NoError(err)
			}
			r.Equal(tc.expectedHvnID, hvnID)
			r.Equal(tc.expectedPeeringID, peeringID)

		})
	}
}
