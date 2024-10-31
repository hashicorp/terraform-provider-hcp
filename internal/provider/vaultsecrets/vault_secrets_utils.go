package vaultsecrets

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type Provider string

const (
	ProviderAWS          Provider = "aws"
	ProviderGCP          Provider = "gcp"
	ProviderMongoDBAtlas Provider = "mongodb_atlas"
	ProviderTwilio       Provider = "twilio"
	ProviderConfluent    Provider = "confluent"
	ProviderPostgres     Provider = "postgres"
)

func (p Provider) String() string {
	return string(p)
}

var (
	secretNameValidator = stringvalidator.RegexMatches(regexp.MustCompile(`^[_\da-zA-Z]{3,36}$`),
		"must contain only letters, numbers or underscores",
	)
	slugValidator = stringvalidator.RegexMatches(regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`),
		"must contain only letters, numbers or hyphens",
	)
)

// locationAttributes are attribute shared between all Vault Secrets resources
var locationAttributes = map[string]schema.Attribute{
	"organization_id": schema.StringAttribute{
		Description: "HCP organization ID that owns the HCP Vault Secrets integration.",
		Computed:    true,
	},
	"project_id": schema.StringAttribute{
		Description: "HCP project ID that owns the HCP Vault Secrets integration. Inferred from the provider configuration if omitted.",
		Computed:    true,
		Optional:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	},
}

// sharedIntegrationAttributes are the attributes shared between all the Vault Secrets integrations
var sharedIntegrationAttributes = map[string]schema.Attribute{
	"resource_id": schema.StringAttribute{
		Description: "Resource ID used to uniquely identify the integration instance on the HCP platform.",
		Computed:    true,
	},
	"resource_name": schema.StringAttribute{
		Description: "Resource name used to uniquely identify the integration instance on the HCP platform.",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The Vault Secrets integration name.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			slugValidator,
		},
	},
	"capabilities": schema.SetAttribute{
		ElementType: types.StringType,
		Description: "Capabilities enabled for the integration. See the Vault Secrets documentation for the list of supported capabilities per provider.",
		Required:    true,
	},
}

// managedSecretAttributes are attribute shared between secrets with a managed lifecycle, such as Dynamic or Rotating
var managedSecretAttributes = map[string]schema.Attribute{
	"app_name": schema.StringAttribute{
		Description: "Vault Secrets application name that owns the secret.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			slugValidator,
		},
	},
	"secret_provider": schema.StringAttribute{
		Description: "The third party platform the dynamic credentials give access to. One of `aws` or `gcp`.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The Vault Secrets secret name.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			secretNameValidator,
		},
	},
	"integration_name": schema.StringAttribute{
		Description: "The Vault Secrets integration name with the capability to manage the secret's lifecycle.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			slugValidator,
		},
	},
}

// resourceFunc is used to get the appropriate Terraform Vault Secrets integration representation either from the plan (create, update) or the state (read, delete)
type resourceFunc func(ctx context.Context, target interface{}) diag.Diagnostics

// operationFunc performs the desired operation (read, create, update, delete) on the Vault Secrets backend
type operationFunc func(r hvsResource) (any, error)

// hvsResource abstracts the conversion between Terraform and HVS domains
type hvsResource interface {
	projectID() types.String
	initModel(ctx context.Context, orgID, projID string) diag.Diagnostics
	fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics
}

// decorateOperation abstracts all the conversion between the Terraform and HVS domain,
// as well as all the statefile management when performing operations (read, create, update, delete)
func decorateOperation[T hvsResource](ctx context.Context, c *clients.Client, state *tfsdk.State, resourceFunc resourceFunc, operation string, operationFunc operationFunc) diag.Diagnostics {
	diags := diag.Diagnostics{}

	var concreteIntegration T
	diags.Append(resourceFunc(ctx, &concreteIntegration)...)
	if diags.HasError() {
		return diags
	}

	orgID, projID := c.Location(concreteIntegration.projectID())
	diags.Append(concreteIntegration.initModel(ctx, orgID, projID)...)
	if diags.HasError() {
		return diags
	}

	model, err := operationFunc(concreteIntegration)
	if err != nil {
		diags.AddError(fmt.Sprintf("Error %s Vault Secrets resource", operation), err.Error())
		return diags
	}
	if model == nil {
		state.RemoveResource(ctx)
		return diags
	}

	diags.Append(concreteIntegration.fromModel(ctx, orgID, projID, model)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(state.Set(ctx, &concreteIntegration)...)

	return diags
}
