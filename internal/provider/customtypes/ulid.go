package customtypes

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/oklog/ulid"
)

// ULIDType is a custom type for ULIDs
type ULIDType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = &ULIDType{}
var _ xattr.TypeWithValidate = ULIDType{}

func (t ULIDType) String() string {
	return "ULIDType"
}

func (t ULIDType) ValueType(context.Context) attr.Value {
	return ULIDValue{}
}

func (t ULIDType) Equal(o attr.Type) bool {
	other, ok := o.(ULIDType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t ULIDType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return ULIDValue{
		StringValue: in,
	}, nil
}

func (t ULIDType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	ulidValue, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to ULIDValue: %v", diags)
	}

	return ulidValue, nil
}

func NewULIDValue(value string) ULIDValue {
	return ULIDValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// ULIDValue is a custom value used to validate that a string is a ULID
type ULIDValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = ULIDValue{}

func (v ULIDValue) Type(context.Context) attr.Type {
	return ULIDType{}
}

func (v ULIDValue) Equal(o attr.Value) bool {
	other, ok := o.(ULIDValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v ULIDValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ULIDValue)
	if !ok {
		diags.Append(newSemanticEqualityCheckTypeError[basetypes.StringValuable](v, newValuable))
		return false, diags
	}

	oldULID, err := ulid.Parse(v.ValueString())
	if err != nil {
		diags.AddError("expected old value to be a valid ULID", err.Error())
	}
	newULID, err := ulid.Parse(newValue.ValueString())
	if err != nil {
		diags.AddError("expected new value to be a valid ULID", err.Error())
	}

	if diags.HasError() {
		return false, diags
	}

	return reflect.DeepEqual(oldULID, newULID), diags
}

// Validate checks that the value is a valid ULID, if it is known and not null.
func (t ULIDType) Validate(ctx context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var valueString string

	if err := value.As(&valueString); err != nil {
		diags.Append(newInvalidTerraformValueError(valuePath, err))
		return diags
	}

	if _, err := ulid.Parse(valueString); err != nil {
		diags.AddAttributeError(
			valuePath,
			"expected a valid ULID",
			err.Error(),
		)
		return diags
	}

	return diags
}
