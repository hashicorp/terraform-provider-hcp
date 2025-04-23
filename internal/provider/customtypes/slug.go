// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/input"
)

var (
	_ basetypes.StringTypable = &SlugType{}
)

type SlugType struct {
	basetypes.StringType
}

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

var (
	_ basetypes.StringValuable       = &SlugValue{}
	_ xattr.ValidateableAttribute    = &SlugValue{}
	_ function.ValidateableParameter = &SlugValue{}
)

type SlugValue struct {
	basetypes.StringValue
}

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

func (v SlugValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if !input.IsSlug(v.ValueString()) {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"expected a valid Slug",
				"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
			),
		)
	}
}

func (v SlugValue) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	if !input.IsSlug(v.ValueString()) {
		resp.Error = function.NewArgumentFuncError(
			req.Position,
			"slugs must be of length 3-36 and contain only alphanumeric characters and hyphens",
		)
	}
}

func NewSlugValue(value string) SlugValue {
	return SlugValue{
		StringValue: basetypes.NewStringValue(value),
	}
}
