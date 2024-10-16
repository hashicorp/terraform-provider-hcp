// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webhookvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ validator.List = uniqueSubscriptionsValidator{}

// uniqueSubscriptionsValidator implements the validator.
type uniqueSubscriptionsValidator struct{}

// UniqueSubscriptions returns a validator which ensures that the configured subscriptions are unique
func UniqueSubscriptions() validator.List {
	return uniqueSubscriptionsValidator{}
}

// Description returns the plaintext description of the validator.
func (v uniqueSubscriptionsValidator) Description(_ context.Context) string {
	return "subscriptions must be unique"
}

// MarkdownDescription returns the Markdown description of the validator.
func (v uniqueSubscriptionsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList implements the validation logic.
func (v uniqueSubscriptionsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	subscriptions := req.ConfigValue.Elements()
	resourceIDMap := map[string]struct{}{}
	for _, subscription := range subscriptions {
		subscriptionValuable, ok := subscription.(basetypes.ObjectValuable)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid subscription events",
				"While performing schema-based validation, an unexpected error occurred. "+
					"The attribute declares an Object value validator, however its values do not implement types.ObjectType interface for custom Object types. "+
					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Path: %s\n", req.Path.String())+
					fmt.Sprintf("Element Type: %T\n", SubscriptionType)+
					fmt.Sprintf("Element Value Type: %T\n", subscription),
			)
			return
		}

		eventsValue, diags := subscriptionValuable.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		// Only return early if the new diagnostics indicate an issue since
		// it likely will be the same for all elements.
		if diags.HasError() {
			return
		}

		elements := eventsValue.Attributes()
		for att, element := range elements {
			if att == "events" {
				validateEvents(ctx, element, req, resp)
				continue
			}

			elementValuable, ok := element.(basetypes.StringValuable)
			if !ok {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid subscription resource id",
					"While performing schema-based validation, an unexpected error occurred. "+
						"The attribute declares a String values validator, however its values do not implement types.StringType or the types.StringTypable interface for custom String types. "+
						"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
						fmt.Sprintf("Path: %s\n", req.Path.String())+
						fmt.Sprintf("Element Type: %T\n", types.StringType)+
						fmt.Sprintf("Element Value Type: %T\n", subscription),
				)
				return
			}

			elementValue, diags := elementValuable.ToStringValue(ctx)
			resp.Diagnostics.Append(diags...)

			// Only return early if the new diagnostics indicate an issue since
			// it likely will be the same for all elements.
			if diags.HasError() {
				return
			}

			resourceID := elementValue.ValueString()
			if _, ok := resourceIDMap[resourceID]; ok {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
					req.Path,
					"duplicated subscription resource_id found. "+
						"Only one subscription per resource_id is allowed",
					elementValue.ValueString(),
				))
			}

			resourceIDMap[resourceID] = struct{}{}
		}
	}
}

func validateEvents(ctx context.Context, events attr.Value, req validator.ListRequest, resp *validator.ListResponse) {
	eventsValuable, ok := events.(basetypes.ListValue)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid subscription events",
			"While performing schema-based validation, an unexpected error occurred. "+
				"The attribute declares a List value validator, however its values do not implement types.ListType interface for custom List types. "+
				"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Path: %s\n", req.Path.String())+
				fmt.Sprintf("Element Type: %T\n", EventsType)+
				fmt.Sprintf("Element Value Type: %T\n", events),
		)
		return
	}

	elements := eventsValuable.Elements()
	sourceMap := map[string]struct{}{}
	for _, event := range elements {
		eventValuable, ok := event.(basetypes.ObjectValuable)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid subscription events",
				"While performing schema-based validation, an unexpected error occurred. "+
					"The attribute declares an Object value validator, however its values do not implement types.ObjectType interface for custom Object types. "+
					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Path: %s\n", req.Path.String())+
					fmt.Sprintf("Element Type: %T\n", EventType)+
					fmt.Sprintf("Element Value Type: %T\n", event),
			)
			return
		}

		eventsValue, diags := eventValuable.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		// Only return early if the new diagnostics indicate an issue since
		// it likely will be the same for all elements.
		if diags.HasError() {
			return
		}

		eventElements := eventsValue.Attributes()
		for att, element := range eventElements {
			if att == "actions" {
				validateActions(ctx, element, req, resp)
				continue
			}

			elementValuable, ok := element.(basetypes.StringValuable)
			if !ok {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid subscription event action",
					"While performing schema-based validation, an unexpected error occurred. "+
						"The attribute declares a String values validator, however its values do not implement types.StringType or the types.StringTypable interface for custom String types. "+
						"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
						fmt.Sprintf("Path: %s\n", req.Path.String())+
						fmt.Sprintf("Element Type: %T\n", types.StringType)+
						fmt.Sprintf("Element Value Type: %T\n", element),
				)
			}

			elementValue, diags := elementValuable.ToStringValue(ctx)
			resp.Diagnostics.Append(diags...)
			// Only return early if the new diagnostics indicate an issue since
			// it likely will be the same for all elements.
			if diags.HasError() {
				return
			}

			source := elementValue.ValueString()
			if _, ok := sourceMap[source]; ok {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
					req.Path,
					"duplicated subscription event source found. "+
						"The event source should be unique per subscription",
					elementValue.ValueString(),
				))
			}

			sourceMap[source] = struct{}{}
		}
	}

}

func validateActions(ctx context.Context, actions attr.Value, req validator.ListRequest, resp *validator.ListResponse) {
	actionsValuable, ok := actions.(basetypes.ListValuable)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid subscription events",
			"While performing schema-based validation, an unexpected error occurred. "+
				"The attribute declares a List value validator, however its values do not implement types.ListType interface for custom List types. "+
				"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Path: %s\n", req.Path.String())+
				fmt.Sprintf("Element Type: %T\n", ActionType)+
				fmt.Sprintf("Element Value Type: %T\n", actions),
		)
	}

	actionsValue, diags := actionsValuable.ToListValue(ctx)
	resp.Diagnostics.Append(diags...)
	// Only return early if the new diagnostics indicate an issue since
	// it likely will be the same for all elements.
	if diags.HasError() {
		return
	}

	elements := actionsValue.Elements()
	var wildcardFound bool
	wildcard := "*"
	for _, element := range elements {
		elementValuable, ok := element.(basetypes.StringValuable)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid subscription event action",
				"While performing schema-based validation, an unexpected error occurred. "+
					"The attribute declares a String values validator, however its values do not implement types.StringType or the types.StringTypable interface for custom String types. "+
					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Path: %s\n", req.Path.String())+
					fmt.Sprintf("Element Type: %T\n", actionsValue.ElementType(ctx))+
					fmt.Sprintf("Element Value Type: %T\n", element),
			)
			return
		}

		elementValue, diags := elementValuable.ToStringValue(ctx)
		resp.Diagnostics.Append(diags...)
		// Only return early if the new diagnostics indicate an issue since
		// it likely will be the same for all elements.
		if diags.HasError() {
			return
		}

		action := elementValue.ValueString()

		if (action != wildcard && wildcardFound) ||
			(action == wildcard && len(elements) > 1) {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				req.Path,
				"invalid subscription event actions found."+
					"The wildcard '*' subscribes to all event actions and should be set alone",
				req.ConfigValue.String(),
			))
			return
		}

		if action == wildcard {
			wildcardFound = true
		}
	}
}

var ActionType = types.ListType{ElemType: types.StringType}

var EventType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"actions": ActionType,
		"source":  types.StringType,
	},
}

var EventsType = types.ListType{
	ElemType: EventType,
}

var SubscriptionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"resource_id": types.StringType,
		"events":      EventsType,
	},
}
