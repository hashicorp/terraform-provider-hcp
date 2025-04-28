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

func TestSlugValidateAttribute(t *testing.T) {
	testCases := map[string]struct {
		Slug          SlugValue
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			Slug: SlugValue{},
		},
		"valid": {
			Slug: NewSlugValue("SLUG1"),
		},
		"too short": {
			Slug: NewSlugValue("SH"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid Slug",
					"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
				),
			},
		},
		"too long": {
			Slug: NewSlugValue(strings.Repeat("A", 37)),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid Slug",
					"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
				),
			},
		},
		"start with alphanumeric": {
			Slug: NewSlugValue("-SLUG1"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid Slug",
					"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
				),
			},
		},
		"end with alphanumeric": {
			Slug: NewSlugValue("SLUG1-"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"expected a valid Slug",
					"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			testCase.Slug.ValidateAttribute(
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

func TestSlugValidateParameter(t *testing.T) {
	testCases := map[string]struct {
		Slug            SlugValue
		expectedFuncErr *function.FuncError
	}{
		"empty": {
			Slug: SlugValue{},
		},
		"valid": {
			Slug: NewSlugValue("SLUG1"),
		},
		"too short": {
			Slug: NewSlugValue("SH"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
			),
		},
		"too long": {
			Slug: NewSlugValue(strings.Repeat("A", 37)),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
			),
		},
		"start with alphanumeric": {
			Slug: NewSlugValue("-SLUG1"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
			),
		},
		"end with alphanumeric": {
			Slug: NewSlugValue("SLUG1-"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := function.ValidateParameterResponse{}

			testCase.Slug.ValidateParameter(
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
