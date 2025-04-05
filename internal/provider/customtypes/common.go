// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func newSemanticEqualityCheckTypeError[V attr.Value](expected V, got V) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Semantic Equality Check Type Error",
		"An unexpected value type was received while performing semantic equality checks. "+
			"Please report this to the provider developers.\n\n"+
			"Expected Value Type: "+fmt.Sprintf("%T", expected)+"\n"+
			"Got Value Type: "+fmt.Sprintf("%T", got),
	)
}
