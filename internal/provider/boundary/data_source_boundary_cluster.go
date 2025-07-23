package boundary

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &dataSource{}

type dataSource struct {
}

type dataSourceModel struct {
	// ExampleAttribute types.String `tfsdk:"example_attribute"`
	// ID               types.String `tfsdk:"id"`
}

func NewBoundaryClusterDataSource() datasource.DataSource {
	return &dataSource{}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

}

func (ds *dataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "boundary_cluster")
}

func (ds *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the Boundary cluster",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 36),
					stringvalidator.RegexMatches(
						regexp.MustCompile(regexpClusterName),
						"must be between 3 and 36 characters in length and contains only letters, numbers or hyphens",
					),
				},
			},
			// "project_id": schema.StringAttribute{
			// 	Computed: true,
			// },
		},
	}
}
