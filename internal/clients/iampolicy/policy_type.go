package iampolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PolicyDataType is a custom type for handling marshaled policy_data.
type PolicyDataType struct {
	basetypes.StringType
}

func (t PolicyDataType) Equal(o attr.Type) bool {
	other, ok := o.(PolicyDataType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t PolicyDataType) String() string {
	return "PolicyDataType"
}

func (t PolicyDataType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := PolicyDataValue{
		StringValue: in,
	}

	return value, nil
}

func (t PolicyDataType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t PolicyDataType) ValueType(ctx context.Context) attr.Value {
	return PolicyDataValue{}
}

// Validate will be called whenever a PolicyDataValue is being created. This is
// helpful to both give the user a better error message but also so the Value
// type can assume the policy_data is valid.
func (t PolicyDataType) Validate(ctx context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var valueString string

	if err := value.As(&valueString); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid Terraform Value",
			"An unexpected error occurred while attempting to convert a Terraform value to a string. "+
				"This generally is an issue with the provider schema implementation. "+
				"Please contact the provider developers.\n\n"+
				"Path: "+valuePath.String()+"\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	var p models.HashicorpCloudResourcemanagerPolicy
	if err := p.UnmarshalBinary([]byte(valueString)); err != nil {
		diags.AddError("failed to unmarshal policy_data", err.Error())
		diags.AddAttributeError(
			valuePath,
			"Invalid IAM Policy Data String Value",
			"An unexpected error occurred while converting a string value that was expected to be a policy data string emitted by a hcp_iam_policy datasource."+
				"Path: "+valuePath.String()+"\n"+
				"Given Value: "+valueString+"\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	return diags
}

type PolicyDataValue struct {
	basetypes.StringValue
}

func (v PolicyDataValue) Equal(o attr.Value) bool {
	other, ok := o.(PolicyDataValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v PolicyDataValue) Type(ctx context.Context) attr.Type {
	return PolicyDataType{}
}

// StringSemanticEquals checks that two policies are semantically equal. This is
// critical for suppressing planned changes where the only delta is the ordering
// of bindings or members within a binding.
func (v PolicyDataValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(PolicyDataValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	// Skipping error checking since the type has a validation which will be
	// called for each Value.
	var existingPolicy models.HashicorpCloudResourcemanagerPolicy
	var newPolicy models.HashicorpCloudResourcemanagerPolicy
	_ = existingPolicy.UnmarshalBinary([]byte(v.ValueString()))
	_ = newPolicy.UnmarshalBinary([]byte(newValue.ValueString()))

	// If the times are equivalent, keep the prior value
	return Equal(&existingPolicy, &newPolicy), diags

}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringTypable = PolicyDataType{}
var _ basetypes.StringValuableWithSemanticEquals = PolicyDataValue{}
