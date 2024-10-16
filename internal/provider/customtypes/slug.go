// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/input"
)

// SlugType is a custom type for Slugs
type SlugType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = &SlugType{}
var _ xattr.TypeWithValidate = SlugType{}

func (t SlugType) String() string {
	return "SlugType"
}

func (t SlugType) ValueType(context.Context) attr.Value {
	return SlugValue{}
}

func (t SlugType) Equal(o attr.Type) bool {
	other, ok := o.(SlugType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t SlugType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return SlugValue{
		StringValue: in,
	}, nil
}

func (t SlugType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	slugValue, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to SlugValue: %v", diags)
	}

	return slugValue, nil
}

func NewSlugValue(value string) SlugValue {
	return SlugValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// SlugValue is a custom value used to validate that a string is a Slug
type SlugValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuable = SlugValue{}

func (v SlugValue) Type(context.Context) attr.Type {
	return SlugType{}
}

func (v SlugValue) Equal(o attr.Value) bool {
	other, ok := o.(SlugValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// Validate checks that the value is a valid Slug, if it is known and not null.
func (t SlugType) Validate(ctx context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var valueString string

	if err := value.As(&valueString); err != nil {
		diags.Append(newInvalidTerraformValueError(valuePath, err))
		return diags
	}

	if !input.IsSlug(valueString) {
		diags.AddAttributeError(
			valuePath,
			"expected a valid Slug",
			"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
		)
		return diags
	}

	return diags
}
