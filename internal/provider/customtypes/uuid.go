// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = &UUIDType{}
)

// UUIDType is a custom type for UUIDs
type UUIDType struct {
	basetypes.StringType
}

func (t UUIDType) String() string {
	return "UUIDType"
}

func (t UUIDType) ValueType(context.Context) attr.Value {
	return UUIDValue{}
}

func (t UUIDType) Equal(o attr.Type) bool {
	other, ok := o.(UUIDType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t UUIDType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return UUIDValue{
		StringValue: in,
	}, nil
}

func (t UUIDType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	uuidValue, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to UUIDValue: %v", diags)
	}

	return uuidValue, nil
}

var (
	_ basetypes.StringValuableWithSemanticEquals = &UUIDValue{}
	_ xattr.ValidateableAttribute                = &UUIDValue{}
	_ function.ValidateableParameter             = &UUIDValue{}
)

type UUIDValue struct {
	basetypes.StringValue
}

func (v UUIDValue) Type(context.Context) attr.Type {
	return UUIDType{}
}

func (v UUIDValue) Equal(o attr.Value) bool {
	other, ok := o.(UUIDValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v UUIDValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(UUIDValue)
	if !ok {
		diags.Append(newSemanticEqualityCheckTypeError[basetypes.StringValuable](v, newValuable))
		return false, diags
	}

	oldUUID, err := uuid.ParseUUID(v.ValueString())
	if err != nil {
		diags.AddError("expected old value to be a valid UUID", err.Error())
	}
	newUUID, err := uuid.ParseUUID(newValue.ValueString())
	if err != nil {
		diags.AddError("expected new value to be a valid UUID", err.Error())
	}

	if diags.HasError() {
		return false, diags
	}

	return reflect.DeepEqual(oldUUID, newUUID), diags
}

func (v UUIDValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if _, err := uuid.ParseUUID(v.ValueString()); err != nil {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"expected a valid UUID",
				err.Error(),
			),
		)
	}
}

func (v UUIDValue) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if _, err := uuid.ParseUUID(v.ValueString()); err != nil {
		resp.Error = function.NewArgumentFuncError(
			req.Position,
			err.Error(),
		)
	}
}

func NewUUIDValue(value string) UUIDValue {
	return UUIDValue{
		StringValue: basetypes.NewStringValue(value),
	}
}
