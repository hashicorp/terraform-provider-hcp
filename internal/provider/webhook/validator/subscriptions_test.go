// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webhookvalidator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	webhookvalidator "github.com/hashicorp/terraform-provider-hcp/internal/provider/webhook/validator"
)

func TestURLValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         func() (basetypes.ListValue, diag.Diagnostics)
		expectError bool
		errMsg      string
	}
	tests := map[string]testCase{
		"unknown String": {
			val: func() (basetypes.ListValue, diag.Diagnostics) {
				return types.ListUnknown(types.ObjectType{}), nil
			},
		},
		"null String": {
			val: func() (basetypes.ListValue, diag.Diagnostics) {
				return types.ListNull(types.ObjectType{}), nil
			},
		},
		"duplicated resource id": {
			val: func() (list basetypes.ListValue, diags diag.Diagnostics) {
				actions, dia := types.ListValue(types.StringType, []attr.Value{
					types.StringValue("*"),
				})
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				eventsOne, dia := types.ObjectValue(
					webhookvalidator.EventType.AttrTypes,
					map[string]attr.Value{
						"actions": actions,
						"source":  types.StringValue("hashicorp.packer.version"),
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				events, dia := types.ListValue(
					webhookvalidator.EventType,
					[]attr.Value{eventsOne},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				subscriptionOne, dia := types.ObjectValue(
					webhookvalidator.SubscriptionType.AttrTypes,
					map[string]attr.Value{
						"resource_id": types.StringValue("someid"),
						"events":      events,
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				subscriptionTwo, dia := types.ObjectValue(
					webhookvalidator.SubscriptionType.AttrTypes,
					map[string]attr.Value{
						"resource_id": types.StringValue("someid"),
						"events":      events,
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				list, dia = types.ListValue(webhookvalidator.SubscriptionType,
					[]attr.Value{
						subscriptionOne,
						subscriptionTwo,
					})
				diags.Append(dia...)
				return
			},
			expectError: true,
			errMsg:      "duplicated subscription resource_id found.",
		},
		"duplicated source": {
			val: func() (list basetypes.ListValue, diags diag.Diagnostics) {
				actions, dia := types.ListValue(types.StringType, []attr.Value{
					types.StringValue("*"),
				})
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				eventOne, dia := types.ObjectValue(
					webhookvalidator.EventType.AttrTypes,
					map[string]attr.Value{
						"actions": actions,
						"source":  types.StringValue("hashicorp.packer.version"),
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				eventTwo, dia := types.ObjectValue(
					webhookvalidator.EventType.AttrTypes,
					map[string]attr.Value{
						"actions": actions,
						"source":  types.StringValue("hashicorp.packer.version"),
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				events, dia := types.ListValue(
					webhookvalidator.EventType,
					[]attr.Value{eventOne, eventTwo},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				subscriptionOne, dia := types.ObjectValue(
					webhookvalidator.SubscriptionType.AttrTypes,
					map[string]attr.Value{
						"resource_id": types.StringValue("someid"),
						"events":      events,
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				list, dia = types.ListValue(webhookvalidator.SubscriptionType,
					[]attr.Value{subscriptionOne})
				diags.Append(dia...)
				return
			},
			expectError: true,
			errMsg:      "duplicated subscription event source found.",
		},
		"multiple actions with '*'": {
			val: func() (list basetypes.ListValue, diags diag.Diagnostics) {
				actions, dia := types.ListValue(types.StringType, []attr.Value{
					types.StringValue("*"),
					types.StringValue("create"),
				})
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				eventOne, dia := types.ObjectValue(
					webhookvalidator.EventType.AttrTypes,
					map[string]attr.Value{
						"actions": actions,
						"source":  types.StringValue("hashicorp.packer.version"),
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				events, dia := types.ListValue(
					webhookvalidator.EventType,
					[]attr.Value{eventOne},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				subscriptionOne, dia := types.ObjectValue(
					webhookvalidator.SubscriptionType.AttrTypes,
					map[string]attr.Value{
						"resource_id": types.StringValue("someid"),
						"events":      events,
					},
				)
				diags.Append(dia...)
				if diags.HasError() {
					return
				}

				list, dia = types.ListValue(webhookvalidator.SubscriptionType,
					[]attr.Value{subscriptionOne})
				diags.Append(dia...)
				return
			},
			expectError: true,
			errMsg:      "invalid subscription event actions found.",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			configVal, diags := test.val()
			if diags.HasError() {
				t.Fatalf("got unexpected error: %s", diags.Errors())
			}

			request := validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    configVal,
			}
			response := validator.ListResponse{}
			webhookvalidator.UniqueSubscriptions().ValidateList(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}

			if response.Diagnostics.HasError() {
				for _, diagnostic := range response.Diagnostics.Errors() {
					if strings.Contains(diagnostic.Detail(), test.errMsg) {
						return
					}
					t.Fatalf("got unexpected error type: %s", response.Diagnostics)
				}
			}
		})
	}
}
