// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = &PackerFingerprintType{}
)

type PackerFingerprintType struct {
	basetypes.StringType
}

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

var (
	_ basetypes.StringValuable       = &PackerFingerprintValue{}
	_ xattr.ValidateableAttribute    = &PackerFingerprintValue{}
	_ function.ValidateableParameter = &PackerFingerprintValue{}
)

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

func (v PackerFingerprintValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	valueString := v.ValueString()

	if len(valueString) < 1 || len(valueString) > 40 {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"invalid format for an HCP Packer Fingerprint",
				"must be between 1 and 40 characters long, inclusive",
			),
		)
		return
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_\-\.\"]+$`).MatchString(valueString) {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"invalid format for an HCP Packer Fingerprint",
				// TODO: The regex also allows double quotes, and does not check the first or last characters.
				// This is because the v1 HCP Packer API allowed quotes. Once that API is deprecated,
				// and there are no more offending versions, we can add the strict validation.
				"must contain only alphanumeric characters, underscores, dashes, and periods",
			),
		)
	}
}

func (v PackerFingerprintValue) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	valueString := v.ValueString()

	if len(valueString) < 1 || len(valueString) > 40 {
		resp.Error = function.NewArgumentFuncError(
			req.Position,
			"HCP Packer Fingerprint must be between 1 and 40 characters long, inclusive",
		)
		return
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_\-\.\"]+$`).MatchString(valueString) {
		resp.Error = function.NewArgumentFuncError(
			req.Position,
			// TODO: The regex also allows double quotes, and does not check the first or last characters.
			// This is because the v1 HCP Packer API allowed quotes. Once that API is deprecated,
			// and there are no more offending versions, we can add the strict validation.
			"HCP Packer Fingerprint must contain only alphanumeric characters, underscores, dashes, and periods",
		)
	}
}

func NewPackerFingerprintValue(value string) PackerFingerprintValue {
	return PackerFingerprintValue{
		StringValue: basetypes.NewStringValue(value),
	}
}
