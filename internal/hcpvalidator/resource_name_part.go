package hcpvalidator

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	validResourceNamePartRegex = "^[A-Za-z0-9][A-Za-z0-9_.-]*$"
	invalidResourceNamePartErr = "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots"
)

var (
	resourceNamePartRegex                  = regexp.MustCompile(validResourceNamePartRegex)
	_                     validator.String = resourceNamePartValidator{}
)

// resourceNamePartValidator validates that a string Attribute's value is a valid
// resource name.
type resourceNamePartValidator struct {
}

// Description describes the validation in plain text formatting.
func (validator resourceNamePartValidator) Description(_ context.Context) string {
	return invalidResourceNamePartErr
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator resourceNamePartValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (v resourceNamePartValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if !resourceNamePartRegex.MatchString(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

// ResourceNamePart returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a string.
//   - Consists of only characters that are valid for a resource name part
//
// Length is not enforced, and should be validated separately.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func ResourceNamePart() validator.String {
	return resourceNamePartValidator{}
}
