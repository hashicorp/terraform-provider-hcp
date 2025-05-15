package vaultradar

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

type radarAssignmentsResource struct {
	client *clients.Client
}

type assignmentData struct {
	ProjectID   types.String `tfsdk:"project_id"`
	Assignments types.List   `tfsdk:"assignments"`
}

type assignmentElementData struct {
	PrincipalID  types.String `tfsdk:"principal_id"`
	Role         types.String `tfsdk:"role"`
	ResourceURIs types.Set    `tfsdk:"resource_uris"`
}

func NewRadarAssignmentsResource() resource.Resource {
	return &radarAssignmentsResource{}
}

func (r *radarAssignmentsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_radar_assignments"
}

func (r *radarAssignmentsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This terraform resource manages a TODO in Vault Radar.",
		Attributes: map[string]schema.Attribute{
			//"id": schema.StringAttribute{
			//	Computed:    true,
			//	Description: "The ID of this resource.",
			//	PlanModifiers: []planmodifier.String{
			//		stringplanmodifier.UseStateForUnknown(),
			//	},
			//},
			// Optional inputs
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assignments": schema.ListNestedAttribute{
				Description: "Assignments TODO",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"principal_id": schema.StringAttribute{
							Required:    true,
							Description: "The principal to bind to the given role.",
						},
						"role": schema.StringAttribute{
							Required:    true,
							Description: "The role name to bind... TODO",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^roles/.+$`),
									"must reference a role name.",
								),
							},
						},
						"resource_uris": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "The radar resource uris to ... TODO",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func (r *radarAssignmentsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *radarAssignmentsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *radarAssignmentsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assignmentData
	diags := req.Plan.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mock creating.

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	data.ProjectID = types.StringValue(projectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *radarAssignmentsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *radarAssignmentsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *radarAssignmentsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assignmentData
	diags := req.Plan.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mock update.

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	data.ProjectID = types.StringValue(projectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
