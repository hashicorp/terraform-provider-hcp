package vaultradar

import (
	"context"
	"fmt"
	"regexp"

	radar_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/data_source_registration_service"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

var (
	_ resource.Resource              = &sourceGitHubEnterpriseResource{}
	_ resource.ResourceWithConfigure = &sourceGitHubEnterpriseResource{}
)

func NewSourceGitHubEnterpriseResource() resource.Resource {
	return &sourceGitHubEnterpriseResource{}
}

type sourceGitHubEnterpriseResource struct {
	client *clients.Client
}

func (r *sourceGitHubEnterpriseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_radar_source_github_enterprise"
}

func (r *sourceGitHubEnterpriseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Vault Radar resource manages HCP Cloud scans of GitHub Enterprise sources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
			"domain_name": schema.StringAttribute{
				Description: "Fully qualified domain name of the server. (Example: myserver.acme.com)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\S+$`),
						"must not contain white space",
					),
				},
			},
			"github_organization": schema.StringAttribute{
				Description: `GitHub organization Vault Radar will monitor. Example: "octocat" for the org https://yourcodeserver.com/octocat`,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`),
						"must contain only letters, numbers, hyphens, underscores, or periods",
					),
				},
			},
			"token": schema.StringAttribute{
				Description: "GitHub personal access token.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
		},
	}
}

type gitHubEnterpriseSource struct {
	ID                 types.String `tfsdk:"id"`
	DomainName         types.String `tfsdk:"domain_name"`
	GitHubOrganization types.String `tfsdk:"github_organization"`
	Token              types.String `tfsdk:"token"`
	ProjectID          types.String `tfsdk:"project_id"`
}

func (r *sourceGitHubEnterpriseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sourceGitHubEnterpriseResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *sourceGitHubEnterpriseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gitHubEnterpriseSource

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	res, err := clients.OnboardRadarSource(ctx, r.client, projectID, radar_service.OnboardDataSourceBody{
		Type:          "github_enterprise",
		Name:          plan.GitHubOrganization.ValueString(),
		ConnectionURL: plan.DomainName.ValueString(),
		Token:         plan.Token.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Radar source", err.Error())
		return
	}

	plan.ID = types.StringValue(res.GetPayload().ID)
	plan.ProjectID = types.StringValue(projectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Trace(ctx, "Created Radar resource")
}

func (r *sourceGitHubEnterpriseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gitHubEnterpriseSource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
	}

	res, err := clients.GetRadarSource(ctx, r.client, projectID, state.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar source not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar source", err.Error())
		return
	}

	// Resource is marked as deleted on the server.
	if res.GetPayload().Deleted {
		tflog.Info(ctx, "Radar source marked for deletion, removing from state.")
		resp.State.RemoveResource(ctx)
		return
	}

	// The only other state that could change related to this resource is the token, and for obvious reasons we don't
	// return that in the read response. So we don't need to update the state here.
	tflog.Trace(ctx, "Read Radar resource")
}

func (r *sourceGitHubEnterpriseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gitHubEnterpriseSource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
	}

	// Assert resource still exists.
	res, err := clients.GetRadarSource(ctx, r.client, projectID, state.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar source not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar source", err.Error())
		return
	}

	// Resource is already marked as being deleted on the server.
	if res.GetPayload().Deleted {
		tflog.Info(ctx, "Radar source marked for deletion, removing from state.")
		resp.State.RemoveResource(ctx)
		return
	}

	// Offboard the Radar source.
	if err := clients.OffboardRadarSource(ctx, r.client, projectID, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Radar source", err.Error())
		return
	}

	tflog.Trace(ctx, "Deleted Radar resource")
}

func (r *sourceGitHubEnterpriseResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported.
	// Plans to support updating the token will be in a future iteration.
	resp.Diagnostics.AddError("Unexpected provider error", "This is an internal error, please report this issue to the provider developers")
}
