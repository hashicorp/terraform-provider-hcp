package hcpvalidator

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	validDisplayNameRegex = `^[-A-Za-z0-9_\."'!]{1}[-A-Za-z0-9_\.\s"'!]{1,34}[-A-Za-z0-9_\."'!]{1}$`
	invalidDisplayNameErr = "only ASCII letters, numbers, hyphens, spaces, quotes, and exclamation points are allowed and may not start or end with a space"
)

var (
	displayNameRegex                  = regexp.MustCompile(validDisplayNameRegex)
	_                validator.String = displayNameValidator{}
)

// displayNameValidator validates that a string Attribute's value is a valid
// display name.
type displayNameValidator struct {
}

// Description describes the validation in plain text formatting.
func (v displayNameValidator) Description(_ context.Context) string {
	return invalidDisplayNameErr
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v displayNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v displayNameValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if !displayNameRegex.MatchString(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

// DisplayName returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a string.
//   - Consists of only characters that are valid for a display name
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func DisplayName() validator.String {
	return displayNameValidator{}
}
