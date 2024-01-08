package logstreaming

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/client/log_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-log-service/preview/2021-03-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
					},
				},
				Required: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
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

	splunkCloud *SplunkCloudProvider `tfsdk:"-"`
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

// extract extracts the Go values from their Terraform wrapped values.
func (h *HCPLogStreamingDestination) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	h.splunkCloud = &SplunkCloudProvider{}
	diags = h.SplunkCloud.As(ctx, h.splunkCloud, basetypes.ObjectAsOptions{})
	return diags
}

// fromModel encodes the values from a Log Streaming Destination model into the
// Terraform values, such that they can be saved to state.
func (h *HCPLogStreamingDestination) fromModel(ctx context.Context, logSD *models.LogService20210330Destination) diag.Diagnostics {
	var diags diag.Diagnostics
	h.Name = types.StringValue(logSD.Name)
	h.StreamingDestinationID = types.StringValue(logSD.Resource.ID)
	h.SplunkCloud = types.ObjectValueMust(h.SplunkCloud.AttributeTypes(ctx), map[string]attr.Value{
		"endpoint": types.StringValue(logSD.SplunkCloudProvider.HecEndpoint),
		"token":    types.StringValue(logSD.SplunkCloudProvider.Token),
	})
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
	createParams.Body = &models.LogService20210330CreateStreamingDestinationRequest{
		DestinationName: plan.Name.ValueString(),
		SplunkCloudProvider: &models.LogService20210330SplunkCloudProvider{
			HecEndpoint: plan.splunkCloud.HecEndpoint.ValueString(),
			Token:       plan.splunkCloud.Token.ValueString(),
		},
	}

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
