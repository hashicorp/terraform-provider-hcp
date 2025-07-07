// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iampolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockResourceIamUpdater is a mock implementation of ResourceIamUpdater interface
type mockResourceIamUpdater struct {
	mock.Mock
}

// GetResourceIamPolicy mocks getting the current IAM policy for a resource
func (m *mockResourceIamUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	args := m.Called(ctx)
	return args.Get(0).(*models.HashicorpCloudResourcemanagerPolicy), args.Get(1).(diag.Diagnostics)
}

// SetResourceIamPolicy mocks setting the IAM policy for a resource
func (m *mockResourceIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	args := m.Called(ctx, policy)
	return args.Get(0).(*models.HashicorpCloudResourcemanagerPolicy), args.Get(1).(diag.Diagnostics)
}

// GetMutexKey mocks getting the mutex key for a resource
func (m *mockResourceIamUpdater) GetMutexKey() string {
	args := m.Called()
	return args.String(0)
}

// Helper function to create a fake updater function that returns a mockResourceIamUpdater
func fakeUpdaterFuncWithMock(mockUpdater *mockResourceIamUpdater) NewResourceIamUpdaterFunc {
	return func(ctx context.Context, d TerraformResourceData, client *clients.Client) (ResourceIamUpdater, diag.Diagnostics) {
		return mockUpdater, diag.Diagnostics{}
	}
}

// Helper function to create a schema for IAM policy resources
func createIamPolicySchema() schema.Schema {
	return schema.Schema{
		Attributes: basePolicySchema,
	}
}

// rmPolicyOption defines a function type for modifying HashicorpCloudResourcemanagerPolicy
type rmPolicyOption func(*models.HashicorpCloudResourcemanagerPolicy)

func withRmPolicyBinding(roleID string, members ...*models.HashicorpCloudResourcemanagerPolicyBindingMember) rmPolicyOption {
	return func(d *models.HashicorpCloudResourcemanagerPolicy) {
		binding := &models.HashicorpCloudResourcemanagerPolicyBinding{
			RoleID:  roleID,
			Members: members,
		}
		d.Bindings = append(d.Bindings, binding)
	}
}

func withRmPolicyEtag(etag string) rmPolicyOption {
	return func(d *models.HashicorpCloudResourcemanagerPolicy) {
		d.Etag = etag
	}
}

func createRmPolicy(rmOptions ...rmPolicyOption) *models.HashicorpCloudResourcemanagerPolicy {
	policy := &models.HashicorpCloudResourcemanagerPolicy{
		Etag:     "",
		Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{},
	}

	for _, rmOption := range rmOptions {
		rmOption(policy)
	}

	return policy
}

func createIamPolicyValue(rmPolicy *models.HashicorpCloudResourcemanagerPolicy, etag string) tftypes.Value {
	policyData, _ := rmPolicy.MarshalBinary()

	return tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"policy_data": PolicyDataType{}.TerraformType(context.Background()),
				"etag":        tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"policy_data": tftypes.NewValue(PolicyDataType{}.TerraformType(context.Background()), string(policyData)),
			"etag":        tftypes.NewValue(tftypes.String, etag),
		},
	)
}

// testScenario contains configuration for a single test case
type testScenario struct {
	name          string
	planPolicy    *models.HashicorpCloudResourcemanagerPolicy
	planEtag      string
	mockGetPolicy *models.HashicorpCloudResourcemanagerPolicy
	mockGetError  diag.Diagnostics
	mockSetPolicy *models.HashicorpCloudResourcemanagerPolicy
	mockSetError  diag.Diagnostics
	expectedError string
}

// testFixture holds and sets up common state across tests
type testFixture struct {
	schema         schema.Schema
	resourcePolicy *resourcePolicy
	mockUpdater    *mockResourceIamUpdater
}

func newTestFixture() *testFixture {
	mockUpdater := &mockResourceIamUpdater{}
	schema := createIamPolicySchema()
	resourcePolicy := &resourcePolicy{
		updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
	}

	return &testFixture{
		schema:         schema,
		resourcePolicy: resourcePolicy,
		mockUpdater:    mockUpdater,
	}
}

func (f *testFixture) SetupMocks(scenario *testScenario) {
	if scenario.mockGetPolicy != nil || scenario.mockGetError.HasError() {
		f.mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(scenario.mockGetPolicy, scenario.mockGetError)
	}

	if scenario.mockSetPolicy != nil || scenario.mockSetError.HasError() {
		f.mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return(scenario.mockSetPolicy, scenario.mockSetError)
	}
}

func (f *testFixture) CreatePlan(scenario *testScenario) tfsdk.Plan {
	return tfsdk.Plan{
		Raw:    createIamPolicyValue(scenario.planPolicy, scenario.planEtag),
		Schema: f.schema,
	}
}

func (f *testFixture) CreateInitialState() tfsdk.State {
	return tfsdk.State{
		Raw:    createIamPolicyValue(createRmPolicy(), ""),
		Schema: f.schema,
	}
}

func (f *testFixture) AssertResults(t *testing.T, scenario *testScenario, plan tfsdk.Plan, initialState tfsdk.State, respState tfsdk.State, diags diag.Diagnostics) {
	t.Helper()
	if scenario.expectedError != "" {
		require.True(t, diags.HasError(), "expected error but got none")

		found := false
		for _, d := range diags {
			if d.Summary() == scenario.expectedError {
				found = true
				break
			}
		}
		require.True(t, found, "expected error '%s' not found in diagnostics: %v", scenario.expectedError, diags)
		// State should remain unchanged on error
		require.Equal(t, initialState.Raw, respState.Raw, "response state should remain unchanged on error, state: %v", respState.Raw)
	} else {
		require.False(t, diags.HasError(), "expected no errors but got: %v", diags.Errors())
		require.True(t, respState.Raw.Equal(plan.Raw), "response state should match plan, state: %v", respState.Raw)
	}
	f.mockUpdater.AssertExpectations(t)
}

func TestResourceIamPolicy_Update(t *testing.T) {
	testCases := []testScenario{
		{
			name: "Successful update",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag:      "new-etag",
			mockGetPolicy: createRmPolicy(withRmPolicyEtag("existing-etag")),
			mockSetPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("new-etag")),
		},
		{
			name: "SetResourceIamPolicy error preserves original state",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag:      "new-etag",
			mockGetPolicy: createRmPolicy(withRmPolicyEtag("existing-etag")),
			mockSetError: diag.Diagnostics{
				diag.NewErrorDiagnostic("set policy error", "failed to set IAM policy"),
			},
			expectedError: "set policy error",
		},
		{
			name: "GetResourceIamPolicy error preserves original state",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag: "new-etag",
			mockGetError: diag.Diagnostics{
				diag.NewErrorDiagnostic("get policy error", "failed to get existing IAM policy"),
			},
			expectedError: "get policy error",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executeUpdateTest(t, tc)
		})
	}
}

func executeUpdateTest(t *testing.T, scenario testScenario) {
	fixture := newTestFixture()
	fixture.SetupMocks(&scenario)
	plan := fixture.CreatePlan(&scenario)
	initialState := fixture.CreateInitialState()
	req := resource.UpdateRequest{Plan: plan}
	resp := &resource.UpdateResponse{State: initialState}

	fixture.resourcePolicy.Update(context.Background(), req, resp)

	fixture.AssertResults(t, &scenario, plan, initialState, resp.State, resp.Diagnostics)
}

func TestResourceIamPolicy_Create(t *testing.T) {
	testCases := []testScenario{
		{
			name: "Successful create",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag:      "new-etag",
			mockGetPolicy: createRmPolicy(withRmPolicyEtag("existing-etag")),
			mockSetPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("new-etag")),
		},
		{
			name: "SetResourceIamPolicy error preserves original state",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag:      "new-etag",
			mockGetPolicy: createRmPolicy(withRmPolicyEtag("existing-etag")),
			mockSetError: diag.Diagnostics{
				diag.NewErrorDiagnostic("set policy error", "failed to set IAM policy"),
			},
			expectedError: "set policy error",
		},
		{
			name: "GetResourceIamPolicy error preserves original state",
			planPolicy: createRmPolicy(withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}), withRmPolicyEtag("")),
			planEtag: "new-etag",
			mockGetError: diag.Diagnostics{
				diag.NewErrorDiagnostic("get policy error", "failed to get existing IAM policy"),
			},
			expectedError: "get policy error",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executeCreateTest(t, tc)
		})
	}
}

func executeCreateTest(t *testing.T, scenario testScenario) {
	fixture := newTestFixture()
	fixture.SetupMocks(&scenario)
	plan := fixture.CreatePlan(&scenario)
	initialState := fixture.CreateInitialState()
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: initialState}

	fixture.resourcePolicy.Create(context.Background(), req, resp)

	fixture.AssertResults(t, &scenario, plan, initialState, resp.State, resp.Diagnostics)
}
