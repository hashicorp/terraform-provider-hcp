// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	crmclient "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// ---- normalizeConstraints ----

func TestNormalizeConstraints_SortsAlphabetically(t *testing.T) {
	input := []string{"constraints/z", "constraints/a", "constraints/m"}
	got := normalizeConstraints(input)
	assert.Equal(t, []string{"constraints/a", "constraints/m", "constraints/z"}, got)
}

func TestNormalizeConstraints_AlreadySorted_Unchanged(t *testing.T) {
	input := []string{"constraints/a", "constraints/b", "constraints/c"}
	got := normalizeConstraints(input)
	assert.Equal(t, input, got)
}

func TestNormalizeConstraints_SingleElement(t *testing.T) {
	input := []string{"constraints/only"}
	got := normalizeConstraints(input)
	assert.Equal(t, []string{"constraints/only"}, got)
}

func TestNormalizeConstraints_EmptySlice(t *testing.T) {
	got := normalizeConstraints([]string{})
	assert.Equal(t, []string{}, got)
}

func TestNormalizeConstraints_NilSlice(t *testing.T) {
	got := normalizeConstraints(nil)
	assert.Equal(t, []string{}, got)
}

func TestNormalizeConstraints_DoesNotMutateInput(t *testing.T) {
	input := []string{"constraints/b", "constraints/a"}
	original := []string{"constraints/b", "constraints/a"}
	_ = normalizeConstraints(input)
	assert.Equal(t, original, input, "normalizeConstraints should not mutate the input slice")
}

func TestNormalizeConstraints_DuplicatesPreservedAndSorted(t *testing.T) {
	input := []string{"constraints/b", "constraints/a", "constraints/b"}
	got := normalizeConstraints(input)
	assert.Equal(t, []string{"constraints/a", "constraints/b", "constraints/b"}, got)
}

// ---- validateConstraints ----

func TestValidateConstraints_AllValid_NoError(t *testing.T) {
	available := map[string]bool{
		"constraints/a": true,
		"constraints/b": true,
	}
	requested := []string{"constraints/a", "constraints/b"}
	diags := validateConstraints(requested, available)
	assert.False(t, diags.HasError())
}

func TestValidateConstraints_OneUnknown_ReturnsError(t *testing.T) {
	available := map[string]bool{
		"constraints/a": true,
	}
	requested := []string{"constraints/a", "constraints/unknown"}
	diags := validateConstraints(requested, available)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "constraints/unknown")
}

func TestValidateConstraints_AllUnknown_ReturnsOneErrorPerUnknown(t *testing.T) {
	available := map[string]bool{}
	requested := []string{"constraints/x", "constraints/y"}
	diags := validateConstraints(requested, available)
	assert.Equal(t, 2, diags.ErrorsCount(), "expected one error per unrecognized constraint")
}

func TestValidateConstraints_EmptyRequested_NoError(t *testing.T) {
	available := map[string]bool{"constraints/a": true}
	diags := validateConstraints([]string{}, available)
	assert.False(t, diags.HasError())
}

func TestValidateConstraints_EmptyRequested_EmptyAvailable_NoError(t *testing.T) {
	diags := validateConstraints([]string{}, map[string]bool{})
	assert.False(t, diags.HasError())
}

func TestValidateConstraints_EmptyStringID_ReturnsError(t *testing.T) {
	available := map[string]bool{"constraints/a": true}
	requested := []string{"constraints/a", ""}
	diags := validateConstraints(requested, available)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "empty")
}

func TestValidateConstraints_ErrorMessageContainsConstraintID(t *testing.T) {
	available := map[string]bool{"constraints/real": true}
	requested := []string{"constraints/typo-here"}
	diags := validateConstraints(requested, available)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "constraints/typo-here")
}

func TestValidateConstraints_ErrorMessageSuggestsListConstraints(t *testing.T) {
	available := map[string]bool{}
	requested := []string{"constraints/bad"}
	diags := validateConstraints(requested, available)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "ListConstraints")
}

// ---- policyToModel ----

func TestPolicyToModel_ConstraintsAreSorted(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: []string{"constraints/z", "constraints/a", "constraints/m"},
		Etag:               "etag-abc",
		OrganizationID:     "org-123",
	}

	model := policyToModel(policy)

	var got []string
	diags := model.EnabledConstraints.ElementsAs(context.Background(), &got, false)
	require.False(t, diags.HasError())
	assert.ElementsMatch(t, []string{"constraints/a", "constraints/m", "constraints/z"}, got)
}

func TestPolicyToModel_IDMatchesOrgID(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: []string{},
		OrganizationID:     "org-456",
		Etag:               "etag-xyz",
	}

	model := policyToModel(policy)
	assert.Equal(t, "org-456", model.ID.ValueString())
	assert.Equal(t, "org-456", model.OrganizationID.ValueString())
}

func TestPolicyToModel_EtagPreserved(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: []string{},
		OrganizationID:     "org-1",
		Etag:               "etag-12345",
	}
	model := policyToModel(policy)
	assert.Equal(t, "etag-12345", model.Etag.ValueString())
}

func TestPolicyToModel_EmptyConstraints_ProducesEmptyList(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: []string{},
		OrganizationID:     "org-1",
		Etag:               "",
	}
	model := policyToModel(policy)

	var got []string
	diags := model.EnabledConstraints.ElementsAs(context.Background(), &got, false)
	require.False(t, diags.HasError())
	assert.Empty(t, got)
}

func TestPolicyToModel_NilConstraints_ProducesEmptyList(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: nil,
		OrganizationID:     "org-1",
		Etag:               "",
	}
	model := policyToModel(policy)

	var got []string
	diags := model.EnabledConstraints.ElementsAs(context.Background(), &got, false)
	require.False(t, diags.HasError())
	assert.Empty(t, got)
}

func TestPolicyToModel_SingleConstraint(t *testing.T) {
	policy := &models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse{
		EnabledConstraints: []string{"constraints/only"},
		OrganizationID:     "org-1",
		Etag:               "etag-1",
	}
	model := policyToModel(policy)
	var got []string
	diags := model.EnabledConstraints.ElementsAs(context.Background(), &got, false)
	require.False(t, diags.HasError())
	assert.ElementsMatch(t, []string{"constraints/only"}, got)
}

// ---- stringSetFromTF ----

func TestStringSetFromTF_ReturnsCorrectStrings(t *testing.T) {
	set, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"a", "b", "c"})
	require.False(t, diags.HasError())

	got, d := stringSetFromTF(context.Background(), set)
	require.False(t, d.HasError())
	assert.ElementsMatch(t, []string{"a", "b", "c"}, got)
}

func TestStringSetFromTF_NullSet_ReturnsEmpty(t *testing.T) {
	set := types.SetNull(types.StringType)
	got, d := stringSetFromTF(context.Background(), set)
	require.False(t, d.HasError())
	assert.Empty(t, got)
}

func TestStringSetFromTF_UnknownSet_ReturnsEmpty(t *testing.T) {
	set := types.SetUnknown(types.StringType)
	got, d := stringSetFromTF(context.Background(), set)
	require.False(t, d.HasError())
	assert.Empty(t, got)
}

func TestStringSetFromTF_EmptySet_ReturnsEmpty(t *testing.T) {
	set, diags := types.SetValueFrom(context.Background(), types.StringType, []string{})
	require.False(t, diags.HasError())
	got, d := stringSetFromTF(context.Background(), set)
	require.False(t, d.HasError())
	assert.Empty(t, got)
}

// ---- setPolicy ----

func TestSetPolicy_SendsEtagWhenAvailable(t *testing.T) {
	var setCalls int32
	var requestBody models.HashicorpCloudResourcemanagerOrganizationServiceSetResourceControlPolicyBody

	r := newTestResourceControlPolicyResource(t, func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPut, req.Method)
		assert.Equal(t, "/resource-manager/2019-12-10/organizations/org-123/resource-control-policy", req.URL.Path)
		atomic.AddInt32(&setCalls, 1)

		require.NoError(t, json.NewDecoder(req.Body).Decode(&requestBody))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(models.HashicorpCloudResourcemanagerSetResourceControlPolicyResponse{Etag: "etag-next"}))
	})

	diags := r.setPolicy(context.Background(), "org-123", []string{"constraints/a"}, "etag-123")
	require.False(t, diags.HasError())
	assert.Equal(t, int32(1), atomic.LoadInt32(&setCalls))
	assert.Equal(t, "etag-123", requestBody.Etag)
	assert.Equal(t, []string{"constraints/a"}, requestBody.ConstraintIds)
}

func TestSetPolicy_ConflictReturnsDiagnosticAndDoesNotRetry(t *testing.T) {
	testCases := map[string]struct {
		httpStatus int
		rpcCode    codes.Code
	}{
		"aborted": {
			httpStatus: http.StatusConflict,
			rpcCode:    codes.Aborted,
		},
		"failed precondition": {
			httpStatus: http.StatusPreconditionFailed,
			rpcCode:    codes.FailedPrecondition,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var setCalls int32

			r := newTestResourceControlPolicyResource(t, func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodPut, req.Method)
				atomic.AddInt32(&setCalls, 1)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.httpStatus)
				require.NoError(t, json.NewEncoder(w).Encode(sharedmodels.GoogleRPCStatus{
					Code:    int32(tc.rpcCode),
					Message: "requested policy conflicts with persisted policy for this scope",
				}))
			})

			diags := r.setPolicy(context.Background(), "org-123", []string{"constraints/a"}, "etag-123")
			require.True(t, diags.HasError())
			assert.Equal(t, int32(1), atomic.LoadInt32(&setCalls))
			assert.Equal(t, "Conflict setting organization resource control policy", diags[0].Summary())
			assert.Contains(t, diags[0].Detail(), "terraform apply")
			assert.Contains(t, diags[0].Detail(), tc.rpcCode.String())
			assert.Contains(t, diags[0].Detail(), "requested policy conflicts")
		})
	}
}

func TestSetPolicy_NonConflictFailureReturnsDiagnosticAndDoesNotRetry(t *testing.T) {
	var setCalls int32

	r := newTestResourceControlPolicyResource(t, func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPut, req.Method)
		atomic.AddInt32(&setCalls, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		require.NoError(t, json.NewEncoder(w).Encode(sharedmodels.GoogleRPCStatus{
			Code:    int32(codes.Internal),
			Message: "internal service failure",
		}))
	})

	diags := r.setPolicy(context.Background(), "org-123", []string{"constraints/a"}, "etag-123")
	require.True(t, diags.HasError())
	assert.Equal(t, int32(1), atomic.LoadInt32(&setCalls))
	assert.Equal(t, "Failed to set organization resource control policy", diags[0].Summary())
	assert.Contains(t, diags[0].Detail(), "terraform apply")
	assert.Contains(t, diags[0].Detail(), codes.Internal.String())
	assert.Contains(t, diags[0].Detail(), "internal service failure")
}

// ---- resource struct satisfies interface ----

func TestResourceOrganizationResourceControlPolicy_ImplementsResource(t *testing.T) {
	var _ resource.Resource = &resourceOrganizationResourceControlPolicy{}
	var _ resource.ResourceWithConfigure = &resourceOrganizationResourceControlPolicy{}
}

func TestNewOrganizationResourceControlPolicyResource_NotNil(t *testing.T) {
	r := NewOrganizationResourceControlPolicyResource()
	assert.NotNil(t, r)
}

// ---- schema ----

func TestResourceSchema_HasExpectedAttributes(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "organization_id")
	assert.Contains(t, attrs, "enabled_constraints")
	assert.Contains(t, attrs, "etag")
}

func TestResourceSchema_OrganizationIDRequiresReplace(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	orgIDAttr, ok := resp.Schema.Attributes["organization_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, orgIDAttr.Required)
	assert.NotEmpty(t, orgIDAttr.PlanModifiers)
}

func TestResourceSchema_EtagIsComputed(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	etagAttr, ok := resp.Schema.Attributes["etag"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, etagAttr.Computed)
	assert.False(t, etagAttr.Required)
	assert.False(t, etagAttr.Optional)
}

func TestResourceSchema_EnabledConstraintsIsRequired(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	attr, ok := resp.Schema.Attributes["enabled_constraints"].(schema.SetAttribute)
	require.True(t, ok)
	assert.True(t, attr.Required)
}

// ---- configure ----

func TestConfigure_NilProviderData_NoError(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestConfigure_WrongProviderDataType_ReturnsError(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "not-a-client"}, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestConfigure_CorrectProviderData_SetsClient(t *testing.T) {
	r := &resourceOrganizationResourceControlPolicy{}
	resp := &resource.ConfigureResponse{}
	c := &clients.Client{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: c}, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, c, r.client)
}

func newTestResourceControlPolicyResource(t *testing.T, handler http.HandlerFunc) *resourceOrganizationResourceControlPolicy {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	api := crmclient.NewHTTPClientWithConfig(nil, crmclient.DefaultTransportConfig().
		WithHost(serverURL.Host).
		WithSchemes([]string{serverURL.Scheme}).
		WithBasePath("/"))

	return &resourceOrganizationResourceControlPolicy{
		client: &clients.Client{Organization: api.OrganizationService},
	}
}
