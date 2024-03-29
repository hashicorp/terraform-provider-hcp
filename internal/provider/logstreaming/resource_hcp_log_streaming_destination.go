// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logstreaming

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/log_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

const TFProviderSourceChannel = "TERRAFORM"

func NewHCPLogStreamingDestinationResource() resource.Resource {
	return &resourceHCPLogStreamingDestination{}
}

type resourceHCPLogStreamingDestination struct {
	client *clients.Client
}

func (r *resourceHCPLogStreamingDestination) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_streaming_destination"
}

func (r *resourceHCPLogStreamingDestination) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Streaming Destination resource allows users to configure an external log system to stream HCP logs to.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The HCP Log Streaming Destination’s name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 30),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"streaming_destination_id": schema.StringAttribute{
				Description: "The ID of the HCP Log Streaming Destination",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"splunk_cloud": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Description: "The Splunk Cloud endpoint to send logs to.",
						Required:    true,
					},
					"token": schema.StringAttribute{
						Description: "The authentication token that will be used by the platform to access Splunk Cloud.",
						Required:    true,
						Sensitive:   true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Optional: true,
				Validators: []validator.Object{
					// Validate only this attribute or cloudwatch is configured.
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("cloudwatch"),
					}...),
				},
			},
			"cloudwatch": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"external_id": schema.StringAttribute{
						Description: "The external_id to provide when assuming the aws IAM role.",
						Sensitive:   true,
						Required:    true,
					},
					"role_arn": schema.StringAttribute{
						Description: "The role_arn that will be assumed to stream logs.",
						Required:    true,
					},
					"region": schema.StringAttribute{
						Description: "The region the CloudWatch destination is set up to stream to.",
						Required:    true,
					},
					"log_group_name": schema.StringAttribute{
						Description: "The log_group_name of the CloudWatch destination.",
						Optional:    true,
						Computed:    true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Optional: true,
				Validators: []validator.Object{
					// Validate only this attribute or splunk_cloud is configured.
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("splunk_cloud"),
					}...),
				},
			},
		},
	}
}

func (r *resourceHCPLogStreamingDestination) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

type HCPLogStreamingDestination struct {
	Name                   types.String `tfsdk:"name"`
	StreamingDestinationID types.String `tfsdk:"streaming_destination_id"`
	SplunkCloud            types.Object `tfsdk:"splunk_cloud"`
	CloudWatch             types.Object `tfsdk:"cloudwatch"`

	splunkCloud *SplunkCloudProvider `tfsdk:"-"`
	cloudwatch  *CloudWatchProvider  `tfsdk:"-"`
}

type SplunkCloudProvider struct {
	HecEndpoint types.String `tfsdk:"endpoint"`
	Token       types.String `tfsdk:"token"`
}

func (s SplunkCloudProvider) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoint": types.StringType,
		"token":    types.StringType,
	}
}

type CloudWatchProvider struct {
	Region       types.String `tfsdk:"region"`
	ExternalID   types.String `tfsdk:"external_id"`
	RoleArn      types.String `tfsdk:"role_arn"`
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func (s CloudWatchProvider) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"region":         types.StringType,
		"external_id":    types.StringType,
		"role_arn":       types.StringType,
		"log_group_name": types.StringType,
	}
}

// extract extracts the Go values from their Terraform wrapped values.
func (h *HCPLogStreamingDestination) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if !h.SplunkCloud.IsNull() {
		h.splunkCloud = &SplunkCloudProvider{}
		diags = h.SplunkCloud.As(ctx, h.splunkCloud, basetypes.ObjectAsOptions{})
	}

	if !h.CloudWatch.IsNull() {
		h.cloudwatch = &CloudWatchProvider{}
		diags = h.CloudWatch.As(ctx, h.cloudwatch, basetypes.ObjectAsOptions{})
	}

	return diags
}

// fromModel encodes the values from a Log Streaming Destination model into the
// Terraform values, such that they can be saved to state.
func (h *HCPLogStreamingDestination) fromModel(ctx context.Context, logSD *models.LogService20210330Destination) diag.Diagnostics {
	var diags diag.Diagnostics
	h.Name = types.StringValue(logSD.Name)
	h.StreamingDestinationID = types.StringValue(logSD.Resource.ID)
	if logSD.CloudwatchlogsProvider != nil {
		h.CloudWatch = types.ObjectValueMust(h.CloudWatch.AttributeTypes(ctx), map[string]attr.Value{
			"external_id":    types.StringValue(logSD.CloudwatchlogsProvider.ExternalID),
			"region":         types.StringValue(logSD.CloudwatchlogsProvider.Region),
			"role_arn":       types.StringValue(logSD.CloudwatchlogsProvider.RoleArn),
			"log_group_name": types.StringValue(logSD.CloudwatchlogsProvider.LogGroupName),
		})
	}

	if logSD.SplunkCloudProvider != nil {
		h.SplunkCloud = types.ObjectValueMust(h.SplunkCloud.AttributeTypes(ctx), map[string]attr.Value{
			"endpoint": types.StringValue(logSD.SplunkCloudProvider.HecEndpoint),
			"token":    types.StringValue(logSD.SplunkCloudProvider.Token),
		})
	}

	return diags
}

func (r *resourceHCPLogStreamingDestination) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HCPLogStreamingDestination
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.extract(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	createParams := log_service.NewLogServiceCreateStreamingDestinationParams()
	createParams.Context = ctx
	createParams.LocationOrganizationID = loc.OrganizationID
	createParams.LocationProjectID = loc.ProjectID

	createRequestBody := &models.LogService20210330CreateStreamingDestinationRequest{
		DestinationName: plan.Name.ValueString(),
		SourceChannel:   TFProviderSourceChannel,
	}

	if plan.splunkCloud != nil {
		createRequestBody.SplunkCloudProvider = &models.LogService20210330SplunkCloudProvider{
			HecEndpoint: plan.splunkCloud.HecEndpoint.ValueString(),
			Token:       plan.splunkCloud.Token.ValueString(),
		}
	}

	if plan.cloudwatch != nil {
		createRequestBody.CloudwatchlogsProvider = &models.LogService20210330CloudwatchLogsProvider{
			ExternalID:   plan.cloudwatch.ExternalID.ValueString(),
			Region:       plan.cloudwatch.Region.ValueString(),
			RoleArn:      plan.cloudwatch.RoleArn.ValueString(),
			LogGroupName: plan.cloudwatch.LogGroupName.ValueString(),
		}
	}

	createParams.Body = createRequestBody

	res, err := r.client.LogService.LogServiceCreateStreamingDestination(createParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Log Streaming Destination", err.Error())
		return
	}

	err = clients.CreateLogStreamingDestinationOrgFilter(ctx, r.client, loc.OrganizationID, res.GetPayload().Destination.Resource.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Log Streaming Destination Filter", err.Error())
		return
	}

	logStreamingDest, err := clients.GetLogStreamingDestination(ctx, r.client, loc, res.GetPayload().Destination.Resource.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving newly created Log Streaming Destination", err.Error())
	}

	resp.Diagnostics.Append(plan.fromModel(ctx, logStreamingDest)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceHCPLogStreamingDestination) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HCPLogStreamingDestination
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}
	res, err := clients.GetLogStreamingDestination(ctx, r.client, loc, state.StreamingDestinationID.ValueString())
	if err != nil {
		var getErr *log_service.LogServiceGetStreamingDestinationDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error retrieving Log Streaming Destination", err.Error())
		return
	}

	resp.Diagnostics.Append(state.fromModel(ctx, res)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceHCPLogStreamingDestination) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update is not supported
}
func (r *resourceHCPLogStreamingDestination) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HCPLogStreamingDestination
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	err := clients.DeleteLogStreamingDestination(ctx, r.client, loc, state.StreamingDestinationID.ValueString())
	if err != nil {
		var getErr *log_service.LogServiceGetStreamingDestinationDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error deleting Log Streaming Destination", err.Error())
		return
	}
}
