// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcpvalidator

import (
	"context"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	invalidURLErr      = "must be a valid URL"
	invalidHTTPSURLErr = "must be a valid HTTPS URL"
)

var (
	_ validator.String = urlValidator{}
)

// urlValidator validates that a string Attribute's value is a valid URL.
type urlValidator struct {
	httpsOnly bool
}

// Description describes the validation in plain text formatting.
func (v urlValidator) Description(_ context.Context) string {
	if v.httpsOnly {
		return invalidHTTPSURLErr
	}
	return invalidURLErr
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v urlValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the actual validation.
func (v urlValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if !govalidator.IsURL(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
		return
	}

	if v.httpsOnly {
		urlValue, err := url.ParseRequestURI(value)
		if err != nil || urlValue.Scheme != "https" {
			response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				request.Path,
				v.Description(ctx),
				value,
			))
		}
	}
}

// URL returns an AttributeValidator which ensures that any configured
// attribute value satisfies requirements of a valid URL.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func URL() validator.String {
	return urlValidator{}
}

// HTTPSUrl returns an AttributeValidator which ensures that any configured
// attribute value satisfies requirements of a valid HTTPS URL.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func HTTPSUrl() validator.String {
	return urlValidator{
		httpsOnly: true,
	}
}
