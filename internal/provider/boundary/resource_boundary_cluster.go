package boundary

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	regexpClusterName = `^[\da-zA-Z][-a-zA-Z\d]{1,34}[\da-zA-Z]$`
)

var (
	_ resource.Resource = &boundaryClusterResource{}
	// _ resource.ResourceWithConfigure = &boundaryClusterResource{}
)

type boundaryClusterResource struct {
}

func NewBoundaryClusterResource() resource.Resource {
	return &boundaryClusterResource{}
}

func (r *boundaryClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boundary_cluster"
}

func (r *boundaryClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// required inputs
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the Boundary cluster",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},

				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 36),
					stringvalidator.RegexMatches(
						regexp.MustCompile(regexpClusterName),
						"must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username of the initial admin user. This must be at least 3 characters in length, alphanumeric, hyphen, or period.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[a-z0-9.-]{3,}$"),
						"invalid boundary username; login name must be all-lowercase alphanumeric, period or hyphen, and at least 3 characters.",
					),
				},
			},
			"password": schema.StringAttribute{
				Description: "The password of the initial admin user. This must be at least 8 characters in length. Note that this may show up in logs, and it will be stored in the state file.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(8),
				},
			},
			"tier": schema.StringAttribute{
				Description: "The tier that the HCP Boundary cluster will be provisioned as, 'Standard' or 'Plus'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					// Make it case-insensitive
					// DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// 	return strings.EqualFold(oldValue, newValue)
					// },
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Standard", "Plus"),
				},
			},
			// Optional inputs
			"project_id": schema.StringAttribute{
				Description: `
The ID of the HCP project where the Boundary cluster is located. If not specified, the project configured in the HCP provider config block will be used.
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Optional: true,
				Computed: true,
				// TODO: Add validation for UUID format
			},
			"maintenance_window_config": schema.ListAttribute{
				Description: "The maintenance window configuration for when cluster upgrades can take place.",
				Optional:    true,
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"day":          types.StringType,
						"start":        types.Int32Type,
						"end":          types.Int32Type,
						"upgrade_type": types.StringType,
					},
				},
			},
			"auth_token_time_to_live": schema.StringAttribute{
				Description: "The time to live for the auth token in golang's time.Duration string format.",
				Optional:    true,
				Computed:    true,
			},
			"auth_token_time_to_stale": schema.StringAttribute{
				Description: "The time to stale for the auth token in golang's time.Duration string format.",
				Optional:    true,
				// Default:     "24h0m0s",
				// TODO: Add validation for UUID format
			},
		},
	}
}

func (r *boundaryClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// resp.Schema = boundaryClusterSchema
}

func (r *boundaryClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Implement the logic to create a Boundary cluster resource
}

func (r *boundaryClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Implement the logic to read a Boundary cluster resource
}

func (r *boundaryClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Implement the logic to read a Boundary cluster resource
}

func (r *boundaryClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implement the logic to update a Boundary cluster resource
}
