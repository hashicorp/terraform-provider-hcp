// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func Test_readWorkloadIdentity(t *testing.T) {
	tcs := map[string]struct {
		model     WorkloadIdentityFrameworkModel
		want      clients.ClientConfig
		wantDiags diag.Diagnostics
	}{
		"missing token and file": {
			model: WorkloadIdentityFrameworkModel{
				ResourceName: basetypes.NewStringValue("my_resource"),
			},
			wantDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("invalid workload_identity", "at least one of `token_file` or `token` must be set"),
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
			},
		},
		"token": {
			model: WorkloadIdentityFrameworkModel{
				ResourceName: basetypes.NewStringValue("my_resource"),
				Token:        basetypes.NewStringValue("my_token"),
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityToken:        "my_token",
			},
		},
		"file": {
			model: WorkloadIdentityFrameworkModel{
				ResourceName: basetypes.NewStringValue("my_resource"),
				TokenFile:    basetypes.NewStringValue("/path/to/token/file"),
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
			},
		},
		"both": {
			model: WorkloadIdentityFrameworkModel{
				ResourceName: basetypes.NewStringValue("my_resource"),
				Token:        basetypes.NewStringValue("my_token"),
				TokenFile:    basetypes.NewStringValue("/path/to/token/file"),
			},
			want: clients.ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityToken:        "my_token",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
			},
		},
	}
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			got, gotDiags := readWorkloadIdentity(tc.model, clients.ClientConfig{})
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantDiags, gotDiags); diff != "" {
				t.Errorf("unexpected diagnostics (-want +got):\n%s", diff)
			}
		})
	}
}
