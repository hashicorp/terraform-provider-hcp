// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/auth/workload"
)

func Test_loadCredentialFile(t *testing.T) {
	tcs := map[string]struct {
		config ClientConfig
		want   *auth.CredentialFile
	}{
		"empty config": {
			config: ClientConfig{},
		},
		"load with geography": {
			config: ClientConfig{
				Geography: "eu",
			},
		},
		"ignores with only resource": {
			config: ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
			},
		},
		"loads with file": {
			config: ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
			},
			want: &auth.CredentialFile{
				Scheme: auth.CredentialFileSchemeWorkload,
				Workload: &workload.IdentityProviderConfig{
					ProviderResourceName: "my_resource",
					File: &workload.FileCredentialSource{
						Path: "/path/to/token/file",
					},
				},
			},
		},
		"loads with token": {
			config: ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityToken:        "my_token",
			},
			want: &auth.CredentialFile{
				Scheme: auth.CredentialFileSchemeWorkload,
				Workload: &workload.IdentityProviderConfig{
					ProviderResourceName: "my_resource",
					Token: &workload.CredentialTokenSource{
						Token: "my_token",
					},
				},
			},
		},
		"defaults to token": {
			config: ClientConfig{
				WorkloadIdentityResourceName: "my_resource",
				WorkloadIdentityTokenFile:    "/path/to/token/file",
				WorkloadIdentityToken:        "my_token",
			},
			want: &auth.CredentialFile{
				Scheme: auth.CredentialFileSchemeWorkload,
				Workload: &workload.IdentityProviderConfig{
					ProviderResourceName: "my_resource",
					Token: &workload.CredentialTokenSource{
						Token: "my_token",
					},
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			got := loadCredentialFile(tc.config)
			if cmp.Diff(tc.want, got) != "" {
				t.Errorf("mismatch (-want +got): %s", cmp.Diff(tc.want, got))
			}
		})
	}

}
