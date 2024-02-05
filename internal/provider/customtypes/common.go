// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func newInvalidTerraformValueError(valuePath path.Path, err error) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		valuePath,
		"Invalid Terraform Value",
		"An unexpected error occurred while attempting to convert a Terraform value to a string. "+
			"This generally is an issue with the provider schema implementation. "+
			"Please contact the provider developers.\n\n"+
			"Path: "+valuePath.String()+"\n"+
			"Error: "+err.Error(),
	)
}

func newSemanticEqualityCheckTypeError[V attr.Value](expected V, got V) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Semantic Equality Check Type Error",
		"An unexpected value type was received while performing semantic equality checks. "+
			"Please report this to the provider developers.\n\n"+
			"Expected Value Type: "+fmt.Sprintf("%T", expected)+"\n"+
			"Got Value Type: "+fmt.Sprintf("%T", got),
	)
}
