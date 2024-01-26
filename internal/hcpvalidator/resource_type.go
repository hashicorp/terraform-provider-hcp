// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcpvalidator

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	validResourceTypeRegex = `^hashicorp(\.([a-z][a-z-]*)){2}$`
	invalidResourceTypeErr = "a Resource Type must have format hashicorp.<service>.<resource> and consist only of lowercase alphabetic characters, dashes, and dots"
)

var _ validator.String = resourceTypeValidator{}
var resourceTypeRegex = regexp.MustCompile(validResourceTypeRegex)

// resourceNamePartValidator validates that a string Attribute's value is a valid
// resource type.
type resourceTypeValidator struct {
}

// Description describes the validation in plain text formatting.
func (v resourceTypeValidator) Description(_ context.Context) string {
	return invalidResourceTypeErr
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v resourceTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the actual validation.
func (v resourceTypeValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if !resourceTypeRegex.MatchString(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

// ResourceType returns an AttributeValidator which ensures that any configured
// attribute value satisfies requirements.
// Resource Type must have a format
//
//	<resource type> = hashicorp.<service>.<resource>
//
// where <service> and <resource> may contain lower-case alphabetic characters as well as dashes
// (-).
//
// For example:
//   - hashicorp.vault.cluster
//   - hashicorp.network.hvn
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func ResourceType() validator.String {
	return resourceTypeValidator{}
}
