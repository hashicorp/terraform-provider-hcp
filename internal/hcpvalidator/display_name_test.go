// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcpvalidator_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

func TestDisplayNameValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid String": {
			val: types.StringValue("good_resource_name"),
		},
		"invalid String": {
			val:         types.StringValue(" bad "),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			hcpvalidator.DisplayName().ValidateString(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestDisplayNameValidator_Good(t *testing.T) {
	t.Parallel()

	tests := []string{
		"goodName", "1goodName", "goodName1", "good-Name", "good1Name2", "goodNamE",
		"GOODNAME", "GOODNAME1", "GOOD-NAME", "GOOD1NAME2", "GOODNAME",
		"good---------NAME", "Good Name", "Good \"Name\"", "Good 'Name'",
		"Good Name!!!", "Good - Name - !",
	}

	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    types.StringValue(test),
			}
			response := validator.StringResponse{}
			hcpvalidator.DisplayName().ValidateString(context.TODO(), request, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestDisplayNameValidator_Bad(t *testing.T) {
	t.Parallel()

	tests := []string{
		"bad@", "bad#", "bad$", "bad%", "bad^", "bad&", "bad*", "bad(", "bad)", "bad/", "bad\\",
		"a", "1", "-",
		"ab", "12", "--",
		"toolongggggggggggggggggggggggggggggggg",
		"5Ὂg̀9! ℃ᾭG", "asdad!$@!$&)!$!)",
		" bad ", " bad", " bad ",
	}

	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    types.StringValue(test),
			}
			response := validator.StringResponse{}
			hcpvalidator.DisplayName().ValidateString(context.TODO(), request, &response)
			if !response.Diagnostics.HasError() {
				t.Fatalf("expected error")
			}
		})
	}
}
