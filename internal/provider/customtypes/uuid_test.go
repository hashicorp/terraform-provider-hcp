// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestUUIDValidateAttribute(t *testing.T) {
	testCases := map[string]struct {
		UUID          UUIDValue
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			UUID: UUIDValue{},
		},
		"valid": {
			UUID: NewUUIDValue("08f9eeef-d2ab-498a-ab55-8e5af6fd1d61"),
		},
		"invalid": {
			UUID: NewUUIDValue("not-a-uuid"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid UUID",
					"uuid string is wrong length",
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			testCase.UUID.ValidateAttribute(
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

func TestUUIDValidateParameter(t *testing.T) {
	testCases := map[string]struct {
		UUID            UUIDValue
		expectedFuncErr *function.FuncError
	}{
		"empty": {
			UUID: UUIDValue{},
		},
		"valid": {
			UUID: NewUUIDValue("08f9eeef-d2ab-498a-ab55-8e5af6fd1d61"),
		},
		"invalid": {
			UUID: NewUUIDValue("not-a-uuid"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"uuid string is wrong length",
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := function.ValidateParameterResponse{}

			testCase.UUID.ValidateParameter(
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
