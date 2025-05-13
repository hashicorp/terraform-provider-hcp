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

func TestULIDValidateAttribute(t *testing.T) {
	testCases := map[string]struct {
		ULID          ULIDValue
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			ULID: ULIDValue{},
		},
		"valid": {
			ULID: NewULIDValue("01G65Z755AFWAKHE12NY0CQ9FH"),
		},
		"invalid": {
			ULID: NewULIDValue("not-a-ULID"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid ULID",
					"ulid: bad data size when unmarshaling",
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			testCase.ULID.ValidateAttribute(
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

func TestULIDValidateParameter(t *testing.T) {
	testCases := map[string]struct {
		ULID            ULIDValue
		expectedFuncErr *function.FuncError
	}{
		"empty": {
			ULID: ULIDValue{},
		},
		"valid": {
			ULID: NewULIDValue("01G65Z755AFWAKHE12NY0CQ9FH"),
		},
		"invalid": {
			ULID: NewULIDValue("not-a-ULID"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"ulid: bad data size when unmarshaling",
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := function.ValidateParameterResponse{}

			testCase.ULID.ValidateParameter(
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
