// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourcePackerBucket{}
var _ resource.ResourceWithImportState = &resourcePackerBucket{}
var _ resource.ResourceWithConfigure = &resourcePackerBucket{}
var _ resource.ResourceWithModifyPlan = &resourcePackerBucket{}

func NewPackerBucketResource() resource.Resource {
	return &resourcePackerBucket{}
}

type resourcePackerBucket struct {
	client *clients.Client
}

func (r *resourcePackerBucket) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packer_bucket"
}

func (r *resourcePackerBucket) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Packer Bucket resource allows you to manage a bucket within an active HCP Packer Registry.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The bucket's name.",
				Validators: []validator.String{
					hcpvalidator.ResourceNamePart(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Optional fields
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to create the bucket under. " +
					"If unspecified, the bucket will be created in the project the provider is configured with.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"resource_name": schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("The buckets's HCP resource name in the format `%s`.",
					"packer/project/<project_id>/packer/<name>"),
			},
			"created_at": schema.StringAttribute{
				Description: "The creation time of this bucket",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where this bucket is located.",
				Computed:    true,
			},
		},
	}
}

// This function is required by the interface but should be unreachable
func (r *resourcePackerBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported, the bucket must be re-created to change any user modifable fields
	resp.Diagnostics.AddError("Unexpected provider error", "This is an internal error, please report this issue to the provider developers")

}

func (r *resourcePackerBucket) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *resourcePackerBucket) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

type bucket struct {
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ResourceName   types.String `tfsdk:"resource_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *resourcePackerBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucket

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}
	name := plan.Name.ValueString()
	res, err := packerv2.CreateBucket(ctx, r.client, loc, name)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket", err.Error())
		return
	}

	plan.ResourceName = types.StringValue(res.ResourceName)
	plan.ProjectID = types.StringValue(res.Location.ProjectID)
	plan.OrganizationID = types.StringValue(res.Location.OrganizationID)
	plan.CreatedAt = types.StringValue(res.CreatedAt.String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePackerBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucket

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := packerservice.NewPackerServiceGetBucketParams()
	params.SetLocationOrganizationID(state.OrganizationID.ValueString())
	params.SetLocationProjectID(state.ProjectID.ValueString())

	params.SetBucketName(state.Name.ValueString())
	bucketResp, err := r.client.PackerV2.PackerServiceGetBucket(params, nil)

	if err != nil {
		if getBucketErr, ok := err.(*packerservice.PackerServiceGetBucketDefault); ok {
			if getBucketErr.IsCode(http.StatusNotFound) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error retrieving bucket", err.Error())
		}
	}
	readBucket := bucketResp.Payload.Bucket
	state.ResourceName = types.StringValue(readBucket.ResourceName)
	state.Name = types.StringValue(readBucket.Name)
	state.CreatedAt = types.StringValue(readBucket.CreatedAt.String())
	state.ProjectID = types.StringValue(readBucket.Location.ProjectID)
	state.OrganizationID = types.StringValue(readBucket.Location.OrganizationID)

	// Save updated state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePackerBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucket
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := packerservice.NewPackerServiceDeleteBucketParams()
	params.SetLocationOrganizationID(state.OrganizationID.ValueString())
	params.SetLocationProjectID(state.ProjectID.ValueString())
	params.SetBucketName(state.Name.ValueString())
	_, err := r.client.PackerV2.PackerServiceDeleteBucket(params, nil)
	if err != nil {
		var getErr *packerservice.PackerServiceDeleteBucketDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting bucket", err.Error())
		return
	}
}

func (r *resourcePackerBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save Resource Name to the State
	resource.ImportStatePassthroughID(ctx, path.Root("resource_name"), req, resp)
	resourceName := req.ID
	resourceNameRegex := `packer\/project\/.*\/bucket\/.*`
	// packer/project/{project_id}/bucket/{bucket_name}

	orgID := r.client.Config.OrganizationID
	resourceParts := strings.SplitN(resourceName, "/", 5)

	resourceNameValid, err := regexp.MatchString(resourceNameRegex, resourceName)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider error", "Failed to parse regular expression, this is an internal error, please report this issue to the provider developers")
		return
	}
	if !resourceNameValid || len(resourceParts) != 5 {
		resp.Diagnostics.AddError("Invalid resource name", "Resource name expected to match packer/project/{project_id}/bucket/{bucket_name}")
		return
	}
	projectID := resourceParts[2]
	bucketName := resourceParts[4]
	resp.State.SetAttribute(ctx, path.Root("name"), bucketName)
	resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)
	resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)
}
