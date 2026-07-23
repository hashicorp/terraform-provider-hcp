// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"google.golang.org/grpc/codes"
)

// resourceControlPolicyModel is the Terraform state/plan model for hcp_resource_control_policy.
type resourceControlPolicyModel struct {
	ID                 types.String `tfsdk:"id"`
	OrganizationID     types.String `tfsdk:"organization_id"`
	EnabledConstraints types.Set    `tfsdk:"enabled_constraints"`
	Etag               types.String `tfsdk:"etag"`
}

// NewOrganizationResourceControlPolicyResource creates a new resource instance.
func NewOrganizationResourceControlPolicyResource() resource.Resource {
	return &resourceOrganizationResourceControlPolicy{}
}

type resourceOrganizationResourceControlPolicy struct {
	client *clients.Client
}

var _ resource.Resource = &resourceOrganizationResourceControlPolicy{}
var _ resource.ResourceWithConfigure = &resourceOrganizationResourceControlPolicy{}

func (r *resourceOrganizationResourceControlPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_control_policy"
}

func (r *resourceOrganizationResourceControlPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the organization-level resource control policy constraints for an HCP organization. " +
			"Enabled constraints define restrictions applied to resources within the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the resource. Set to the organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the organization to manage the resource control policy for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled_constraints": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The list of constraint IDs to enable for the organization. " +
					"Each constraint ID must be a recognized constraint returned by ListConstraints.",
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "The etag of the current resource control policy. Used internally for concurrency control.",
			},
		},
	}
}

func (r *resourceOrganizationResourceControlPolicy) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

// Create implements resource.Resource.
func (r *resourceOrganizationResourceControlPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceControlPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrganizationID.ValueString()

	// Extract desired constraint IDs from plan.
	desiredConstraints, diags := stringSetFromTF(ctx, plan.EnabledConstraints)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Call ListConstraints and validate all desired constraints are recognized.
	availableConstraints, diags := r.listAllConstraints(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateConstraints(desiredConstraints, availableConstraints)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 2: Use the current policy etag when one already exists.
	currentPolicy, diags := r.getPolicy(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	etag := ""
	if currentPolicy != nil {
		etag = currentPolicy.Etag
	}

	setDiags := r.setPolicy(ctx, orgID, desiredConstraints, etag)
	resp.Diagnostics.Append(setDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 3: Read-after-write to get the latest state including etag.
	policy, diags := r.getPolicy(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 4: Map policy into state.
	state := policyToModel(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read implements resource.Resource.
func (r *resourceOrganizationResourceControlPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceControlPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.OrganizationID.ValueString()

	policy, diags := r.getPolicy(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If policy not found, remove from state.
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := policyToModel(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *resourceOrganizationResourceControlPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceControlPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state resourceControlPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrganizationID.ValueString()

	desiredConstraints, diags := stringSetFromTF(ctx, plan.EnabledConstraints)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	availableConstraints, diags := r.listAllConstraints(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateConstraints(desiredConstraints, availableConstraints)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setDiags := r.setPolicy(ctx, orgID, desiredConstraints, etagFromState(state))
	resp.Diagnostics.Append(setDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedPolicy, diags := r.getPolicy(ctx, orgID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updatedPolicy == nil {
		resp.Diagnostics.AddError(
			"Failed to read organization resource control policy after update",
			"The policy update succeeded, but the provider could not read the updated policy afterwards. Rerun terraform apply to reconcile state.",
		)
		return
	}

	newState := policyToModel(updatedPolicy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *resourceOrganizationResourceControlPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete not supported in PR1",
		"This initial resource implementation supports create and read only. Please remove the resource from state manually until delete support is added in a follow-up.",
	)
}

// ---- helpers ----

// listAllConstraints calls ListConstraints, paging through all results, and
// returns a map of constraint ID -> true for O(1) lookup.
func (r *resourceOrganizationResourceControlPolicy) listAllConstraints(ctx context.Context, orgID string) (map[string]bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	available := make(map[string]bool)
	var nextPageToken *string

	for {
		params := organization_service.NewOrganizationServiceListConstraintsParamsWithContext(ctx)
		params.ID = orgID
		if nextPageToken != nil {
			params.PaginationNextPageToken = nextPageToken
		}

		res, err := r.client.Organization.OrganizationServiceListConstraints(params, nil)
		if err != nil {
			serviceErr, ok := err.(*organization_service.OrganizationServiceListConstraintsDefault)
			if !ok {
				diags.AddError("Failed to list organization constraints", err.Error())
				return nil, diags
			}
			diags.AddError(
				"Failed to list organization constraints",
				fmt.Sprintf("API error %d: %s", serviceErr.Code(), err.Error()),
			)
			return nil, diags
		}

		payload := res.GetPayload()
		if payload == nil {
			break
		}

		for _, c := range payload.Constraints {
			if c != nil && c.ID != "" {
				available[c.ID] = true
			}
		}

		// Check for next page.
		if payload.Pagination == nil || payload.Pagination.NextPageToken == "" {
			break
		}
		token := payload.Pagination.NextPageToken
		nextPageToken = &token
	}

	return available, diags
}

// setPolicy calls SetResourceControlPolicy with the given constraint IDs and etag.
func (r *resourceOrganizationResourceControlPolicy) setPolicy(ctx context.Context, orgID string, constraintIDs []string, etag string) diag.Diagnostics {
	var diags diag.Diagnostics

	params := organization_service.NewOrganizationServiceSetResourceControlPolicyParamsWithContext(ctx)
	params.ID = orgID
	params.Body = &models.HashicorpCloudResourcemanagerOrganizationServiceSetResourceControlPolicyBody{
		ConstraintIds: constraintIDs,
		Etag:          etag,
	}

	_, err := r.client.Organization.OrganizationServiceSetResourceControlPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*organization_service.OrganizationServiceSetResourceControlPolicyDefault)
		if !ok {
			diags.AddError(
				"Failed to set organization resource control policy",
				fmt.Sprintf("The provider could not update the organization resource control policy. Review the error and rerun terraform apply. Error: %s", err.Error()),
			)
			return diags
		}

		diags.Append(setPolicyDiagnostic(serviceErr))
	}
	return diags
}

func etagFromState(state resourceControlPolicyModel) string {
	if state.Etag.IsNull() || state.Etag.IsUnknown() {
		return ""
	}
	return state.Etag.ValueString()
}

func setPolicyDiagnostic(serviceErr *organization_service.OrganizationServiceSetResourceControlPolicyDefault) diag.Diagnostic {
	if isConflictSetPolicyError(serviceErr) {
		return diag.NewErrorDiagnostic(
			"Conflict setting organization resource control policy",
			fmt.Sprintf("The organization resource control policy changed since Terraform last read it. Rerun terraform apply to refresh the policy etag and retry the change. API error %d (%s): %s", serviceErr.Code(), setPolicyRPCCode(serviceErr), setPolicyErrorMessage(serviceErr)),
		)
	}

	return diag.NewErrorDiagnostic(
		"Failed to set organization resource control policy",
		fmt.Sprintf("The HCP API rejected the organization resource control policy change. Review the API error, fix the underlying issue, and rerun terraform apply. API error %d (%s): %s", serviceErr.Code(), setPolicyRPCCode(serviceErr), setPolicyErrorMessage(serviceErr)),
	)
}

func isConflictSetPolicyError(serviceErr *organization_service.OrganizationServiceSetResourceControlPolicyDefault) bool {
	if serviceErr == nil {
		return false
	}

	payload := serviceErr.GetPayload()
	if payload != nil {
		switch codes.Code(payload.Code) {
		case codes.Aborted, codes.FailedPrecondition:
			return true
		}
	}

	switch serviceErr.Code() {
	case http.StatusConflict, http.StatusPreconditionFailed:
		return true
	default:
		return false
	}
}

func setPolicyRPCCode(serviceErr *organization_service.OrganizationServiceSetResourceControlPolicyDefault) string {
	if serviceErr == nil || serviceErr.GetPayload() == nil {
		return codes.Unknown.String()
	}

	if serviceErr.GetPayload().Code == 0 {
		return codes.Unknown.String()
	}

	return codes.Code(serviceErr.GetPayload().Code).String()
}

func setPolicyErrorMessage(serviceErr *organization_service.OrganizationServiceSetResourceControlPolicyDefault) string {
	if serviceErr == nil {
		return "unknown error"
	}

	if payload := serviceErr.GetPayload(); payload != nil && payload.Message != "" {
		return payload.Message
	}

	return serviceErr.Error()
}

// getPolicy calls GetResourceControlPolicy and returns the response payload.
// Returns nil, nil when the resource is not found (404).
func (r *resourceOrganizationResourceControlPolicy) getPolicy(ctx context.Context, orgID string) (*models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	params := organization_service.NewOrganizationServiceGetResourceControlPolicyParamsWithContext(ctx)
	params.ID = orgID

	res, err := r.client.Organization.OrganizationServiceGetResourceControlPolicy(params, nil)
	if err != nil {
		serviceErr, ok := err.(*organization_service.OrganizationServiceGetResourceControlPolicyDefault)
		if ok && serviceErr.Code() == 404 {
			return nil, diags
		}
		if !ok {
			diags.AddError("Failed to get organization resource control policy", err.Error())
			return nil, diags
		}
		diags.AddError(
			"Failed to get organization resource control policy",
			fmt.Sprintf("API error %d: %s", serviceErr.Code(), err.Error()),
		)
		return nil, diags
	}

	return res.GetPayload(), diags
}

// policyToModel converts an API response into the Terraform state model.
// Constraints are sorted before writing to state for plan stability.
func policyToModel(p *models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse) resourceControlPolicyModel {
	sorted := normalizeConstraints(p.EnabledConstraints)

	constraintVals := make([]types.String, len(sorted))
	for i, c := range sorted {
		constraintVals[i] = types.StringValue(c)
	}

	setVal, _ := types.SetValueFrom(context.Background(), types.StringType, constraintVals)

	return resourceControlPolicyModel{
		ID:                 types.StringValue(p.OrganizationID),
		OrganizationID:     types.StringValue(p.OrganizationID),
		EnabledConstraints: setVal,
		Etag:               types.StringValue(p.Etag),
	}
}

// normalizeConstraints returns a sorted copy of the constraint ID list.
// This ensures plan stability regardless of API return order.
func normalizeConstraints(constraints []string) []string {
	if len(constraints) == 0 {
		return []string{}
	}
	out := make([]string, len(constraints))
	copy(out, constraints)
	sort.Strings(out)
	return out
}

// validateConstraints checks that every requested constraint ID exists in the
// available set returned by ListConstraints. Returns a diagnostic per unknown ID.
func validateConstraints(requested []string, available map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, id := range requested {
		if id == "" {
			diags.AddError(
				"Invalid constraint ID",
				"Constraint ID must not be empty. Remove the empty entry from enabled_constraints.",
			)
			continue
		}
		if !available[id] {
			diags.AddError(
				"Unrecognized constraint ID",
				fmt.Sprintf("Constraint %q is not a recognized constraint for this organization. "+
					"Check the ID for typos or call ListConstraints to see available constraints.", id),
			)
		}
	}
	return diags
}

// stringSetFromTF extracts a []string from a types.Set of strings.
func stringSetFromTF(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return []string{}, diags
	}
	var elems []types.String
	diags.Append(set.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]string, len(elems))
	for i, e := range elems {
		out[i] = e.ValueString()
	}
	return out, diags
}
