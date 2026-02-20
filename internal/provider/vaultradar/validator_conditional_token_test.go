// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestEnvVarRegex tests the EnvVarRegex constant
func TestEnvVarRegex(t *testing.T) {
	t.Parallel()

	regex := regexp.MustCompile(EnvVarRegex)

	validCases := []struct {
		name  string
		value string
	}{
		{name: "uppercase with underscore", value: "GITHUB_TOKEN"},
		{name: "simple uppercase", value: "MY_VAR"},
		{name: "with numbers", value: "TOKEN123"},
		{name: "starts with underscore", value: "_PRIVATE"},
		{name: "short uppercase", value: "ABC"},
		{name: "mixed case with numbers", value: "abc123_XYZ"},
	}

	invalidCases := []struct {
		name  string
		value string
	}{
		{name: "contains hyphen", value: "MY-TOKEN"},
		{name: "contains dot", value: "MY.TOKEN"},
		{name: "contains space", value: "MY TOKEN"},
		{name: "contains at symbol", value: "MY@TOKEN"},
		{name: "empty string", value: ""},
	}

	for _, tc := range validCases {
		t.Run("valid_"+tc.name, func(t *testing.T) {
			if !regex.MatchString(tc.value) {
				t.Errorf("expected %q to be valid but it was invalid", tc.value)
			}
		})
	}

	for _, tc := range invalidCases {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			if regex.MatchString(tc.value) {
				t.Errorf("expected %q to be invalid but it was valid", tc.value)
			}
		})
	}
}

// createTestConfig creates a tfsdk.Config for testing validators
func createTestConfig(detectorType types.String) tfsdk.Config {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token":         schema.StringAttribute{Optional: true},
			"token_env_var": schema.StringAttribute{Optional: true},
			"detector_type": schema.StringAttribute{Optional: true},
		},
	}

	// Build the tftypes value based on whether detector_type is null or has a value
	var detectorTypeValue tftypes.Value
	if detectorType.IsNull() {
		detectorTypeValue = tftypes.NewValue(tftypes.String, nil)
	} else if detectorType.IsUnknown() {
		detectorTypeValue = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	} else {
		detectorTypeValue = tftypes.NewValue(tftypes.String, detectorType.ValueString())
	}

	rawValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"token":         tftypes.String,
				"token_env_var": tftypes.String,
				"detector_type": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"token":         tftypes.NewValue(tftypes.String, nil),
			"token_env_var": tftypes.NewValue(tftypes.String, nil),
			"detector_type": detectorTypeValue,
		},
	)

	return tfsdk.Config{
		Schema: testSchema,
		Raw:    rawValue,
	}
}

func TestTokenRequiredWhen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tokenValue   types.String
		detectorType types.String
		expectError  bool
	}{
		{
			name:         "token provided, detector_type hcp - valid",
			tokenValue:   types.StringValue("ghp_token"),
			detectorType: types.StringValue("hcp"),
			expectError:  false,
		},
		{
			name:         "token provided, detector_type null - valid",
			tokenValue:   types.StringValue("ghp_token"),
			detectorType: types.StringNull(),
			expectError:  false,
		},
		{
			name:         "token null, detector_type hcp - error",
			tokenValue:   types.StringNull(),
			detectorType: types.StringValue("hcp"),
			expectError:  true,
		},
		{
			name:         "token null, detector_type null - error",
			tokenValue:   types.StringNull(),
			detectorType: types.StringNull(),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("token"),
				ConfigValue: tt.tokenValue,
				Config:      createTestConfig(tt.detectorType),
			}

			resp := &validator.StringResponse{}
			TokenRequiredWhen("hcp").ValidateString(context.Background(), req, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error but got: %v", resp.Diagnostics)
			}
		})
	}
}

func TestTokenEnvVarRequiredWhen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tokenEnvVar  types.String
		detectorType types.String
		expectError  bool
	}{
		{
			name:         "token_env_var provided, detector_type agent - valid",
			tokenEnvVar:  types.StringValue("GITHUB_TOKEN"),
			detectorType: types.StringValue("agent"),
			expectError:  false,
		},
		{
			name:         "token_env_var null, detector_type agent - error",
			tokenEnvVar:  types.StringNull(),
			detectorType: types.StringValue("agent"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("token_env_var"),
				ConfigValue: tt.tokenEnvVar,
				Config:      createTestConfig(tt.detectorType),
			}

			resp := &validator.StringResponse{}
			TokenEnvVarRequiredWhen("agent").ValidateString(context.Background(), req, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error but got: %v", resp.Diagnostics)
			}
		})
	}
}

func TestTokenForbiddenWhen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tokenValue   types.String
		detectorType types.String
		expectError  bool
	}{
		{
			name:         "token provided, detector_type agent - error",
			tokenValue:   types.StringValue("ghp_token"),
			detectorType: types.StringValue("agent"),
			expectError:  true,
		},
		{
			name:         "token provided, detector_type hcp - valid",
			tokenValue:   types.StringValue("ghp_token"),
			detectorType: types.StringValue("hcp"),
			expectError:  false,
		},
		{
			name:         "token provided, detector_type null - valid",
			tokenValue:   types.StringValue("ghp_token"),
			detectorType: types.StringNull(),
			expectError:  false,
		},
		{
			name:         "token null, detector_type agent - valid",
			tokenValue:   types.StringNull(),
			detectorType: types.StringValue("agent"),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("token"),
				ConfigValue: tt.tokenValue,
				Config:      createTestConfig(tt.detectorType),
			}

			resp := &validator.StringResponse{}
			TokenForbiddenWhen("agent").ValidateString(context.Background(), req, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error but got: %v", resp.Diagnostics)
			}
		})
	}
}
