// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestPackerFingerprintValidateAttribute(t *testing.T) {
	testCases := map[string]struct {
		PackerFingerprint PackerFingerprintValue
		expectedDiags     diag.Diagnostics
	}{
		"empty": {
			PackerFingerprint: PackerFingerprintValue{},
		},
		"valid": {
			PackerFingerprint: NewPackerFingerprintValue("01JDAW2CWCX1JHXE54XA7FFXHS"),
		},
		"too short": {
			PackerFingerprint: NewPackerFingerprintValue(""),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"invalid format for an HCP Packer Fingerprint",
					"must be between 1 and 40 characters long, inclusive",
				),
			},
		},
		"too long": {
			PackerFingerprint: NewPackerFingerprintValue(strings.Repeat("A", 41)),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"invalid format for an HCP Packer Fingerprint",
					"must be between 1 and 40 characters long, inclusive",
				),
			},
		},
		"invalid string": {
			PackerFingerprint: NewPackerFingerprintValue("$invalid$"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"invalid format for an HCP Packer Fingerprint",
					"must contain only alphanumeric characters, underscores, dashes, and periods",
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			testCase.PackerFingerprint.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{
					Path: path.Root("test"),
				},
				&resp,
			)

			if diff := cmp.Diff(resp.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestPackerFingerprintValidateParameter(t *testing.T) {
	testCases := map[string]struct {
		PackerFingerprint PackerFingerprintValue
		expectedFuncErr   *function.FuncError
	}{
		"empty": {
			PackerFingerprint: PackerFingerprintValue{},
		},
		"valid": {
			PackerFingerprint: NewPackerFingerprintValue("01JDAW2CWCX1JHXE54XA7FFXHS"),
		},
		"too short": {
			PackerFingerprint: NewPackerFingerprintValue(""),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"HCP Packer Fingerprint must be between 1 and 40 characters long, inclusive",
			),
		},
		"too long": {
			PackerFingerprint: NewPackerFingerprintValue(strings.Repeat("A", 41)),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"HCP Packer Fingerprint must be between 1 and 40 characters long, inclusive",
			),
		},
		"invalid string": {
			PackerFingerprint: NewPackerFingerprintValue("$invalid$"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"HCP Packer Fingerprint must contain only alphanumeric characters, underscores, dashes, and periods",
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := function.ValidateParameterResponse{}

			testCase.PackerFingerprint.ValidateParameter(
				context.Background(),
				function.ValidateParameterRequest{
					Position: int64(0),
				},
				&resp,
			)

			if diff := cmp.Diff(resp.Error, testCase.expectedFuncErr); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
