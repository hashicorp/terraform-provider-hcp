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

func TestResourceIamPolicy_Update(t *testing.T) {
	t.Run("Successful update", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})
		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: iamPolicyExpectedEtag, Bindings: rmPolicy.Bindings}, diag.Diagnostics{})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.UpdateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		r.Update(context.Background(), req, resp)

		require.False(t, resp.Diagnostics.HasError(), "expected no diagnostics errors, but got: %v", resp.Diagnostics.Errors())
		require.True(t, resp.State.Raw.Equal(plan.Raw), "state not updated as expected, response state: %v", resp.State.Raw)
		mockUpdater.AssertExpectations(t)
	})
	t.Run("SetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})
		// Mock SetResourceIamPolicy to return an error
		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("set policy error", "failed to set IAM policy"),
			})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.UpdateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		r.Update(context.Background(), req, resp)

		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have an error")
		// Verify the error came from SetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "set policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to have SetResourceIamPolicy error")
		require.Equal(t, initialState.Raw, resp.State.Raw,
			"expected state to remain unchanged when SetResourceIamPolicy fails, response state: %v", resp.State.Raw)
		mockUpdater.AssertExpectations(t)
	})

	t.Run("GetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		// Mock GetResourceIamPolicy to return an error
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("get policy error", "failed to get existing IAM policy"),
			})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.UpdateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		r.Update(context.Background(), req, resp)

		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have an error")
		// Verify the error came from GetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "get policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to have GetResourceIamPolicy error")
		require.Equal(t, initialState.Raw, resp.State.Raw,
			"expected state to remain unchanged when GetResourceIamPolicy fails, resp state: %v", resp.State.Raw)

		mockUpdater.AssertExpectations(t)
	})
}

func TestResourceIamPolicy_Create(t *testing.T) {
	t.Run("Successful create", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})
		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: iamPolicyExpectedEtag, Bindings: rmPolicy.Bindings}, diag.Diagnostics{})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.CreateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.CreateResponse{
			State: initialState,
		}

		r.Create(context.Background(), req, resp)

		require.False(t, resp.Diagnostics.HasError(), "expected no diagnostics errors, but got: %v", resp.Diagnostics.Errors())
		require.True(t, resp.State.Raw.Equal(plan.Raw), "state not updated as expected, response state: %v", resp.State.Raw)
		mockUpdater.AssertExpectations(t)
	})

	t.Run("SetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})
		// Mock SetResourceIamPolicy to return an error
		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("set policy error", "failed to set IAM policy"),
			})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.CreateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.CreateResponse{
			State: initialState,
		}

		r.Create(context.Background(), req, resp)

		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have an error")
		// Verify the error came from SetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "set policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to have SetResourceIamPolicy error")
		require.Equal(t, initialState.Raw, resp.State.Raw,
			"expected state to remain unchanged when SetResourceIamPolicy fails, response state: %v", resp.State.Raw)
		mockUpdater.AssertExpectations(t)
	})

	t.Run("GetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		iamPolicySchema := createIamPolicySchema()
		rmPolicy := createRmPolicy(
			withRmPolicyBinding("roles/viewer", &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   "user:test@example.com",
				MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
			}),
		)
		iamPolicyExpectedEtag := "new-etag"
		plan := tfsdk.Plan{
			Raw:    createIamPolicyValue(rmPolicy, iamPolicyExpectedEtag),
			Schema: iamPolicySchema,
		}
		// Mock GetResourceIamPolicy to return an error
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("get policy error", "failed to get existing IAM policy"),
			})
		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}
		req := resource.CreateRequest{
			Plan: plan,
		}
		initialState := tfsdk.State{
			Raw:    createIamPolicyValue(createRmPolicy(), ""),
			Schema: iamPolicySchema,
		}
		resp := &resource.CreateResponse{
			State: initialState,
		}

		r.Create(context.Background(), req, resp)

		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have an error")
		// Verify the error came from GetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "get policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to have GetResourceIamPolicy error")
		require.Equal(t, initialState.Raw, resp.State.Raw,
			"expected state to remain unchanged when GetResourceIamPolicy fails, response state: %v", resp.State.Raw)
		mockUpdater.AssertExpectations(t)
	})
}
