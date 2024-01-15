// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	webhookservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-webhook/preview/2023-05-31/client/webhook_service"
	webhookmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-webhook/preview/2023-05-31/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceWebhook{}
var _ resource.ResourceWithImportState = &resourceWebhook{}

func NewWebhookResource() resource.Resource {
	return &resourceWebhook{}
}

type resourceWebhook struct {
	client *clients.Client
}

func (r *resourceWebhook) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *resourceWebhook) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The webhook resource manages a HCP webhook, used to notify external systems about a " +
			"project resource's lifecycle events",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The webhook's name.",
				Validators: []validator.String{
					hcpvalidator.ResourceNamePart(),
				},
			},

			"config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The webhook configuration used to deliver event payloads.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required: true,
						MarkdownDescription: `The HTTP or HTTPS destination URL that HCP delivers the event payloads to. 
The destination must be able to use the HCP webhook 
[payload](https://developer.hashicorp.com/hcp/docs/hcp/admin/projects/webhooks#webhook-payload).`,
						Validators: []validator.String{
							hcpvalidator.URL(),
						},
					},
					"hmac_key": schema.StringAttribute{
						Optional: true,
						Description: "The arbitrary secret that HCP uses to sign all its webhook requests. This is a" +
							"write-only field, it is written once and not visible thereafter.",
						Sensitive: true,
					},
				},
			},

			// Optional fields
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to create the webhook under. " +
					"If unspecified, the webhook will be created in the project the provider is configured with. ",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The webhook's description. Descriptions are useful for helping others understand the purpose of the webhook.",
			},

			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Indicates if the webhook should receive payloads for the subscribed events.",
				Default:             booldefault.StaticBool(true),
			},

			"subscriptions": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Set of events to subscribe the webhook to all resources or a specific resource in the project.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_id": schema.StringAttribute{
							Optional: true,
							Description: "Refers to the resource the webhook is subscribed to. " +
								"If not set, the webhook subscribes to the emitted events listed in events for " +
								"any resource in the webhook's project.",
							Validators: nil,
						},
						"events": schema.ListNestedAttribute{
							Required: true,
							Description: "The information about the events of a webhook subscription. " +
								"The service that owns the resource is responsible for maintaining events. " +
								"Refer to the service's webhook documentation for more information.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"actions": schema.ListAttribute{
										ElementType: types.StringType,
										Required:    true,
										Description: "The list of event actions subscribed for the resource type set as the [source](#source). " +
											"For example, `[\"create\", \"update\"]`. " +
											"When the action is '*', it means that the webhook is subscribed to all event actions for the event source. ",
									},
									"source": schema.StringAttribute{
										Required: true,
										Description: "The resource type of the source of the event. For example, `hashicorp.packer.version`. " +
											"Event source might not be the same type as the resource that the webhook is " +
											"subscribed to ([resource_id](#resource_id)) if the event is from a descendant resource. " +
											"For example, webhooks are subscribed to a `hashicorp.packer.registry` and " +
											"receive events for descendent resources such as a `hashicorp.packer.version`.",
										Validators: []validator.String{
											hcpvalidator.ResourceType(),
										},
									},
								},
							},
						},
					},
				},
			},

			// Computed fields
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The webhook's unique identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("The webhooks's resource name in the format `%s`.",
					"webhook/project/<project_id>/geo/us/webhook/<name>"),
			},
		},
	}
}

func (r *resourceWebhook) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type webhook struct {
	ProjectID     types.String          `tfsdk:"project_id"`
	Name          types.String          `tfsdk:"name"`
	Config        *webhookConfig        `tfsdk:"config"`
	Description   types.String          `tfsdk:"description"`
	Enabled       types.Bool            `tfsdk:"enabled"`
	Subscriptions []webhookSubscription `tfsdk:"subscriptions"`
	ResourceID    types.String          `tfsdk:"resource_id"`
	ResourceName  types.String          `tfsdk:"resource_name"`
}

type webhookConfig struct {
	URL     types.String `tfsdk:"url"`
	HmacKey types.String `tfsdk:"hmac_key"`
}

type webhookSubscription struct {
	ResourceID types.String               `tfsdk:"resource_id"`
	Events     []webhookSubscriptionEvent `tfsdk:"events"`
}

type webhookSubscriptionEvent struct {
	Actions []types.String `tfsdk:"actions"`
	Source  types.String   `tfsdk:"source"`
}

func (r *resourceWebhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhook

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectName := fmt.Sprintf("project/%s", r.client.Config.ProjectID)
	if plan.ProjectID.ValueString() != "" {
		projectName = fmt.Sprintf("project/%s", plan.ProjectID.ValueString())
	}

	createParams := webhookservice.NewWebhookServiceCreateWebhookParams()
	createParams.ParentResourceName = projectName
	createParams.Body = &webhookmodels.HashicorpCloudWebhookCreateWebhookRequestBody{
		Config: &webhookmodels.HashicorpCloudWebhookWebhookConfig{
			HmacKey: plan.Config.HmacKey.ValueString(),
			URL:     plan.Config.URL.ValueString(),
		},
		Description: plan.Description.ValueString(),
		Enabled:     plan.Enabled.ValueBoolPointer(),
		Name:        plan.Name.ValueString(),
	}

	subscriptions := make([]*webhookmodels.HashicorpCloudWebhookWebhookSubscription, len(plan.Subscriptions))
	for i, subscription := range plan.Subscriptions {
		newSubscription := &webhookmodels.HashicorpCloudWebhookWebhookSubscription{
			Events:     make([]*webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent, 0),
			ResourceID: subscription.ResourceID.ValueString(),
		}

		for _, event := range subscription.Events {
			for _, action := range event.Actions {
				newSubscription.Events = append(newSubscription.Events, &webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent{
					Action: action.ValueString(),
					Source: event.Source.ValueString(),
				})
			}
		}
		subscriptions[i] = newSubscription
	}

	if len(subscriptions) > 0 {
		createParams.Body.Subscriptions = subscriptions
	}

	res, err := r.client.Webhook.WebhookServiceCreateWebhook(createParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	// Get project id from created webhook
	webhook := res.GetPayload().Webhook
	if webhook == nil {
		resp.Diagnostics.AddError(
			"Unexpected service response",
			"The Create webhook request didn't fail but returned a nil webhook object. "+
				"Report this issue to the provider developers.")
		return
	}

	projectID, err := webhookProjectID(webhook.ResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving service principal parent", err.Error())
	}

	plan.ResourceID = types.StringValue(webhook.ResourceID)
	plan.ResourceName = types.StringValue(webhook.ResourceName)
	plan.Enabled = types.BoolValue(webhook.Enabled)
	plan.ProjectID = types.StringValue(projectID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// webhookProjectID extracts the parent resource name from the webhook resource name
func webhookProjectID(resourceName string) (string, error) {
	err := fmt.Errorf("unexpected format for webhook resource_name: %q", resourceName)
	parts := strings.Split(resourceName, "/")
	if parts[1] != "project" && len(parts) < 3 {
		return "", err
	}
	return parts[2], nil
}

// webhookName extracts the webhook name from the webhook resource name
func webhookName(resourceName string) (string, error) {
	err := fmt.Errorf("unexpected format for webhook resource_name: %q", resourceName)
	parts := strings.Split(resourceName, "/")
	if parts[5] != "webhook" && len(parts) < 7 {
		return "", err
	}
	return parts[6], nil
}

func (r *resourceWebhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhook

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getParams := webhookservice.NewWebhookServiceGetWebhookParams()
	getParams.ResourceName = state.ResourceName.ValueString()

	res, err := r.client.Webhook.WebhookServiceGetWebhook(getParams, nil)
	if err != nil {
		var getErr *webhookservice.WebhookServiceGetWebhookDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error retrieving webhook", err.Error())
		return
	}

	webhook := res.GetPayload().Webhook
	if webhook == nil {
		resp.Diagnostics.AddError(
			"Unexpected service response",
			"The get webhook request didn't fail but returned a nil webhook object. "+
				"Report this issue to the provider developers.")
		return
	}

	// Get parent from created webhook
	projectID, err := webhookProjectID(webhook.ResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving webhook project id", err.Error())
	}

	// Extract webhook name
	webhookName, err := webhookName(webhook.ResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving webhook name", err.Error())
	}

	state.Name = types.StringValue(webhookName)
	state.ResourceID = types.StringValue(webhook.ResourceID)
	state.ResourceName = types.StringValue(webhook.ResourceName)
	state.Enabled = types.BoolValue(webhook.Enabled)
	state.ProjectID = types.StringValue(projectID)
	state.Description = types.StringValue(webhook.Description)
	if state.Config == nil {
		state.Config = &webhookConfig{}
	}
	state.Config.URL = types.StringValue(webhook.Config.URL)

	webhookSubscriptions := make([]webhookSubscription, len(webhook.Subscriptions))
	for i, subscription := range webhook.Subscriptions {
		newSubscription := webhookSubscription{
			Events: make([]webhookSubscriptionEvent, 0),
		}

		if subscription.ResourceID != "" {
			newSubscription.ResourceID = types.StringValue(subscription.ResourceID)
		}

		eventsMap := make(map[types.String][]types.String)
		for _, event := range subscription.Events {
			eventsMap[types.StringValue(event.Source)] = append(eventsMap[types.StringValue(event.Source)],
				types.StringValue(event.Action))
		}

		for source, actions := range eventsMap {
			newSubscription.Events = append(newSubscription.Events, webhookSubscriptionEvent{
				Actions: actions,
				Source:  source,
			})
		}

		webhookSubscriptions[i] = newSubscription
	}
	if len(webhook.Subscriptions) > 0 ||
		len(webhook.Subscriptions) != len(state.Subscriptions) {
		state.Subscriptions = webhookSubscriptions
	}

	// Save updated state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state webhook

	// Read Terraform plan and state into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		updateNameParams := webhookservice.NewWebhookServiceUpdateWebhookNameParams()
		updateNameParams.ResourceName = state.ResourceName.ValueString()
		updateNameParams.Body = &webhookmodels.HashicorpCloudWebhookUpdateWebhookNameRequestBody{
			Name: plan.Name.ValueString(),
		}

		res, err := r.client.Webhook.WebhookServiceUpdateWebhookName(updateNameParams, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error updating webhook name", err.Error())
			return
		}

		webhook := res.GetPayload().Webhook
		if webhook == nil {
			resp.Diagnostics.AddError(
				"Unexpected service response",
				"The update webhook name request didn't fail but returned a nil webhook object. "+
					"Report this issue to the provider developers.")
			return
		}
		state.ResourceName = types.StringValue(webhook.ResourceName)

		defer func() {
			if resp.Diagnostics.HasError() {
				// If it failed to update the resource but the name still got successfully updated,
				// we update the name and resource name in the state.
				// The next apply the new name will no longer be an update.
				state.Name = plan.Name
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			}
		}()
	}

	var updateMaks []string
	if !plan.Config.URL.Equal(state.Config.URL) ||
		!plan.Config.HmacKey.Equal(state.Config.HmacKey) {
		updateMaks = append(updateMaks, "config")
	}
	if !plan.Description.Equal(state.Description) {
		updateMaks = append(updateMaks, "description")
	}
	if !plan.Enabled.Equal(state.Enabled) {
		updateMaks = append(updateMaks, "enabled")
	}

	subscriptions := make([]*webhookmodels.HashicorpCloudWebhookWebhookSubscription, 0)
	if !reflect.DeepEqual(plan.Subscriptions, state.Subscriptions) {
		for _, subscription := range plan.Subscriptions {
			newSubscription := &webhookmodels.HashicorpCloudWebhookWebhookSubscription{
				Events:     make([]*webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent, 0),
				ResourceID: subscription.ResourceID.ValueString(),
			}

			for _, event := range subscription.Events {
				for _, action := range event.Actions {
					newSubscription.Events = append(newSubscription.Events, &webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent{
						Action: action.ValueString(),
						Source: event.Source.ValueString(),
					})
				}
			}

			subscriptions = append(subscriptions, newSubscription)
		}
		updateMaks = append(updateMaks, "subscriptions")
	}

	if len(updateMaks) > 0 {
		updateParams := webhookservice.NewWebhookServiceUpdateWebhookParams()
		updateParams.ResourceName = state.ResourceName.ValueString()
		updateParams.Body = &webhookmodels.HashicorpCloudWebhookUpdateWebhookRequestBody{
			UpdateMask: strings.Join(updateMaks, ","),
			Webhook: &webhookmodels.HashicorpCloudWebhookWebhook{
				Config: &webhookmodels.HashicorpCloudWebhookWebhookConfig{
					HmacKey: plan.Config.HmacKey.ValueString(),
					URL:     plan.Config.URL.ValueString(),
				},
				Description:   plan.Description.ValueString(),
				Enabled:       plan.Enabled.ValueBool(),
				Subscriptions: subscriptions,
			},
		}

		_, err := r.client.Webhook.WebhookServiceUpdateWebhook(updateParams, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error updating webhook", err.Error())
			return
		}
	}

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	// Update computed resource name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx,
		path.Root("resource_name"), state.ResourceName)...)
}

func (r *resourceWebhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhook
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteParams := webhookservice.NewWebhookServiceDeleteWebhookParams()
	deleteParams.ResourceName = state.ResourceName.ValueString()
	_, err := r.client.Webhook.WebhookServiceDeleteWebhook(deleteParams, nil)
	if err != nil {
		var getErr *webhookservice.WebhookServiceDeleteWebhookDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
		return
	}
}

func (r *resourceWebhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_name"), req, resp)
}
