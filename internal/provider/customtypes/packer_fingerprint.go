package customtypes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PackerFingerprintType is a custom type for HCP Packer Fingerprints
type PackerFingerprintType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = &PackerFingerprintType{}
var _ xattr.TypeWithValidate = PackerFingerprintType{}

func (t PackerFingerprintType) String() string {
	return "PackerFingerprintType"
}

func (t PackerFingerprintType) ValueType(context.Context) attr.Value {
	return PackerFingerprintValue{}
}

func (t PackerFingerprintType) Equal(o attr.Type) bool {
	other, ok := o.(PackerFingerprintType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t PackerFingerprintType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return PackerFingerprintValue{
		StringValue: in,
	}, nil
}

func (t PackerFingerprintType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	fingerprintValue, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to PackerFingerprintValue: %v", diags)
	}

	return fingerprintValue, nil
}

func NewPackerFingerprintValue(value string) PackerFingerprintValue {
	return PackerFingerprintValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// PackerFingerprintValue is a custom value used to validate that a string is an HCP Packer Fingerprint
type PackerFingerprintValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuable = PackerFingerprintValue{}

func (v PackerFingerprintValue) Type(context.Context) attr.Type {
	return PackerFingerprintType{}
}

func (v PackerFingerprintValue) Equal(o attr.Value) bool {
	other, ok := o.(PackerFingerprintValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// Validate checks that the value is a valid PackerFingerprint, if it is known and not null.
func (t PackerFingerprintType) Validate(ctx context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var valueString string

	if err := value.As(&valueString); err != nil {
		diags.Append(newInvalidTerraformValueError(valuePath, err))
		return diags
	}

	if len(valueString) < 1 || len(valueString) > 40 {
		diags.AddAttributeError(
			valuePath,
			"invalid format for an HCP Packer Fingerprint",
			"must be between 1 and 40 characters long, inclusive",
		)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_\-\.\"]+$`).MatchString(valueString) {
		diags.AddAttributeError(
			valuePath,
			"invalid format for an HCP Packer Fingerprint",
			// TODO: The regex also allows double quotes, and does not check the first or last characters.
			// This is because the v1 HCP Packer API allowed quotes. Once that API is deprecated,
			// and there are no more offending versions, we can add the strict validation.
			"must contain only alphanumeric characters, underscores, dashes, and periods",
		)
	}

	return diags
}
