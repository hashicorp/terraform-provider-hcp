package iampolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// GetMutexKey mocks getting the mutex key for the resource
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
func createTestSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_data": schema.StringAttribute{
				CustomType:  PolicyDataType{},
				Required:    true,
				Description: "The policy to apply.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "The etag captures the existing state of the policy.",
			},
		},
	}
}

// Helper function to create a plan with expected structure
func createTestPlan(policyData string) tfsdk.Plan {
	testSchema := createTestSchema()
	ctx := context.Background()
	policyDataType := PolicyDataType{}

	// Create the plan value with proper structure
	planValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"policy_data": policyDataType.TerraformType(ctx),
				"etag":        tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"policy_data": tftypes.NewValue(policyDataType.TerraformType(ctx), policyData),
			"etag":        tftypes.NewValue(tftypes.String, ""),
		},
	)

	plan := tfsdk.Plan{
		Raw:    planValue,
		Schema: testSchema,
	}

	return plan
}

// Helper function to create a state with expected structure
func createTestState(policyData, etag string) tfsdk.State {
	testSchema := createTestSchema()
	ctx := context.Background()
	policyDataType := PolicyDataType{}

	stateValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"policy_data": policyDataType.TerraformType(ctx),
				"etag":        tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"policy_data": tftypes.NewValue(policyDataType.TerraformType(ctx), policyData),
			"etag":        tftypes.NewValue(tftypes.String, etag),
		},
	)

	return tfsdk.State{
		Raw:    stateValue,
		Schema: testSchema,
	}
}

func TestResourceIamPolicy_Update(t *testing.T) {
	t.Run("Successful update", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}
		newEtag := "new-etag"

		// Mock successful operations
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})

		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{
				Etag:     newEtag,
				Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{},
			}, diag.Diagnostics{})

		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}

		// Create plan with proper schema
		plan := createTestPlan(`{"etag":"","bindings":[]}`)

		req := resource.UpdateRequest{
			Plan: plan,
		}

		// Set up response with initial state
		initialState := createTestState(`{"etag":"original-etag","bindings":[]}`, "original-etag")
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		// Call Update
		r.Update(context.Background(), req, resp)

		// Assert that there's no error
		require.False(t, resp.Diagnostics.HasError(), "expected no diagnostics errors")

		// Assert that State.Raw WAS updated to the plan's raw value
		var valueMap map[string]tftypes.Value
		err := resp.State.Raw.As(&valueMap)
		if err != nil {
			t.Fatalf("Failed to convert resp.State.Raw to map: %v", err)
		}
		policyDataValue := valueMap["policy_data"]
		etagValue := valueMap["etag"]
		var policyDataString, etagString string

		policyDataValue.As(&policyDataString)
		etagValue.As(&etagString)

		require.Equal(t, `{"bindings":[]}`, policyDataString)
		require.Equal(t, "new-etag", etagString)

		// Verify mock expectations
		mockUpdater.AssertExpectations(t)
	})
	t.Run("SetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}

		// Mock GetResourceIamPolicy to return a valid policy (needed when etag is empty)
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return(&models.HashicorpCloudResourcemanagerPolicy{Etag: "existing-etag"}, diag.Diagnostics{})

		// Mock SetResourceIamPolicy to return an error - this will cause setIamPolicyData to fail
		mockUpdater.On("SetResourceIamPolicy", mock.Anything, mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("set policy error", "failed to set IAM policy"),
			})

		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}

		// Create plan with proper schema
		plan := createTestPlan(`{"etag":"","bindings":[]}`)

		req := resource.UpdateRequest{
			Plan: plan,
		}

		// Set up response with initial state
		initialState := createTestState(`{"etag":"original-etag","bindings":[]}`, "original-etag")
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		// Store the initial state value for comparison
		// initialStateValue := initialState.Raw
		initialStateValue := resp.State.Raw

		// Call Update
		r.Update(context.Background(), req, resp)

		// Assert that there's an error from setIamPolicyData
		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have error")

		// Verify the error came from SetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "set policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to find the SetResourceIamPolicy error")

		// Key assertion: State.Raw was not updated
		require.Equal(t, initialStateValue, resp.State.Raw,
			"expected State.Raw to remain unchanged when setIamPolicyData fails")

		// Verify mock expectations
		mockUpdater.AssertExpectations(t)
	})

	t.Run("GetResourceIamPolicy error preserves original state", func(t *testing.T) {
		mockUpdater := &mockResourceIamUpdater{}

		// Mock GetResourceIamPolicy to return an error
		mockUpdater.On("GetResourceIamPolicy", mock.Anything).
			Return((*models.HashicorpCloudResourcemanagerPolicy)(nil), diag.Diagnostics{
				diag.NewErrorDiagnostic("get policy error", "failed to get existing IAM policy"),
			})

		r := &resourcePolicy{
			updaterFunc: fakeUpdaterFuncWithMock(mockUpdater),
		}

		// Create plan with empty etag - this will trigger GetResourceIamPolicy
		plan := createTestPlan(`{"etag":"","bindings":[]}`)

		req := resource.UpdateRequest{
			Plan: plan,
		}

		// Set up response with initial state
		initialState := createTestState(`{"etag":"original","bindings":[]}`, "original-etag")
		resp := &resource.UpdateResponse{
			State: initialState,
		}

		// Store the initial state value for comparison
		initialStateValue := initialState.Raw

		// Call Update
		r.Update(context.Background(), req, resp)

		// Assert that there's an error from GetResourceIamPolicy
		require.True(t, resp.Diagnostics.HasError(), "expected diagnostics to have error")

		// Verify the error came from GetResourceIamPolicy
		found := false
		for _, diag := range resp.Diagnostics {
			if diag.Summary() == "get policy error" {
				found = true
				break
			}
		}
		require.True(t, found, "expected to find the GetResourceIamPolicy error")

		// Assert that State.Raw was NOT updated (should remain the initial state)
		require.Equal(t, initialStateValue, resp.State.Raw,
			"expected State.Raw to remain unchanged when GetResourceIamPolicy fails")

		// Verify mock expectations
		mockUpdater.AssertExpectations(t)
	})
}
