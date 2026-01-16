// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvVarRegex is the regular expression used to validate environment variable names.
// It allows one or more characters from the set [A–Z, a–z, 0–9, _] and rejects all
// other characters (such as '-', '.', or whitespace). This simplified pattern is
// intentionally more permissive than the strict POSIX convention to support common
// cross-platform usage while still excluding characters that are not widely supported
// in environment variable identifiers.
const EnvVarRegex = `^[a-zA-Z0-9_]+$`

var _ validator.String = tokenRequiredWhenValidator{}

// tokenRequiredWhenValidator validates that token is provided when detector_type matches specified values or is not specified
type tokenRequiredWhenValidator struct {
	detectorTypes []string
}

// TokenRequiredWhen returns a validator which ensures that token is required when detector_type matches the specified values
func TokenRequiredWhen(detectorTypes ...string) validator.String {
	return tokenRequiredWhenValidator{
		detectorTypes: detectorTypes,
	}
}

func (v tokenRequiredWhenValidator) Description(_ context.Context) string {
	return fmt.Sprintf("token is required when detector_type is one of %v or not specified", v.detectorTypes)
}

func (v tokenRequiredWhenValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v tokenRequiredWhenValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If token is provided (known) or will be known later (unknown), validation passes
	if !req.ConfigValue.IsNull() {
		return
	}

	// Get the detector_type value from the config
	var detectorType types.String
	diags := req.Config.GetAttribute(ctx, path.Root("detector_type"), &detectorType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if detector_type is null (defaults to 'hcp') or matches one of the specified types
	requiresToken := detectorType.IsNull()
	if !requiresToken {
		for _, dt := range v.detectorTypes {
			if detectorType.ValueString() == dt {
				requiresToken = true
				break
			}
		}
	}

	// If token is required but not provided, add error (regardless of token_env_var)
	if requiresToken {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			fmt.Sprintf("The attribute %s is required when detector_type is %v or not specified.", req.Path, v.detectorTypes),
		)
	}
}

var _ validator.String = tokenEnvVarRequiredWhenValidator{}

// tokenEnvVarRequiredWhenValidator validates that token_env_var is provided when detector_type matches specified values
type tokenEnvVarRequiredWhenValidator struct {
	detectorTypes []string
}

// TokenEnvVarRequiredWhen returns a validator which ensures that token_env_var is required when detector_type matches the specified values
func TokenEnvVarRequiredWhen(detectorTypes ...string) validator.String {
	return tokenEnvVarRequiredWhenValidator{
		detectorTypes: detectorTypes,
	}
}

func (v tokenEnvVarRequiredWhenValidator) Description(_ context.Context) string {
	return fmt.Sprintf("token_env_var is required when detector_type is one of %v", v.detectorTypes)
}

func (v tokenEnvVarRequiredWhenValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v tokenEnvVarRequiredWhenValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If token_env_var is already provided or will be known later (unknown), validation passes
	if !req.ConfigValue.IsNull() {
		return
	}

	// Get the detector_type value from the config
	var detectorType types.String
	diags := req.Config.GetAttribute(ctx, path.Root("detector_type"), &detectorType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if detector_type matches one of the specified types (e.g., 'agent')
	requiresTokenEnvVar := false
	if !detectorType.IsNull() {
		for _, dt := range v.detectorTypes {
			if detectorType.ValueString() == dt {
				requiresTokenEnvVar = true
				break
			}
		}
	}

	// If token_env_var is required but not provided, add error (regardless of token)
	if requiresTokenEnvVar {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			fmt.Sprintf("The attribute %s is required when detector_type is %v.", req.Path, v.detectorTypes),
		)
	}
}

var _ validator.String = tokenForbiddenWhenValidator{}

// tokenForbiddenWhenValidator validates that token is not provided when detector_type matches specified values
type tokenForbiddenWhenValidator struct {
	detectorTypes []string
}

// TokenForbiddenWhen returns a validator which ensures that token cannot be used when detector_type matches the specified values
func TokenForbiddenWhen(detectorTypes ...string) validator.String {
	return tokenForbiddenWhenValidator{
		detectorTypes: detectorTypes,
	}
}

func (v tokenForbiddenWhenValidator) Description(_ context.Context) string {
	return fmt.Sprintf("token cannot be used when detector_type is one of %v", v.detectorTypes)
}

func (v tokenForbiddenWhenValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v tokenForbiddenWhenValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If token is not provided, no conflict
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get the detector_type value from the config
	var detectorType types.String
	diags := req.Config.GetAttribute(ctx, path.Root("detector_type"), &detectorType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if detector_type matches one of the forbidden types
	isForbidden := false
	if !detectorType.IsNull() {
		for _, dt := range v.detectorTypes {
			if detectorType.ValueString() == dt {
				isForbidden = true
				break
			}
		}
	}

	// If token is forbidden, add error
	if isForbidden {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Attribute Configuration",
			fmt.Sprintf("The attribute token cannot be used when detector_type is %v. Use token_env_var instead.", v.detectorTypes),
		)
	}
}
