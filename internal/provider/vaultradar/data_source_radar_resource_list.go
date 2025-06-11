package vaultradar

import (
	"context"
	"fmt"

	rrs "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/resource_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceRadarResourceList struct {
	client *clients.Client
}

type DataSourceRadarResourceListModel struct {
	URILikeFilter                []types.String `tfsdk:"uri_like_filter"`
	URILikeFilterCaseInsensitive types.Bool     `tfsdk:"uri_like_filter_case_insensitive"`
	Resources                    []Resource     `tfsdk:"resources"`
}

// Resource represents a radar resource in the list.
type Resource struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	URI             types.String `tfsdk:"uri"`
	ConnectionURL   types.String `tfsdk:"connection_url"`
	DataSourceName  types.String `tfsdk:"data_source_name"`
	DataSourceType  types.String `tfsdk:"data_source_type"`
	DataSourceInfo  types.String `tfsdk:"data_source_info"`
	Description     types.String `tfsdk:"description"`
	DetectorType    types.String `tfsdk:"detector_type"`
	Visibility      types.String `tfsdk:"visibility"`
	State           types.String `tfsdk:"state"`
	HCPResourceID   types.String `tfsdk:"hcp_resource_id"`
	HCPResourceName types.String `tfsdk:"hcp_resource_name"`
}

func NewRadarResourceListDataSource() datasource.DataSource {
	return &DataSourceRadarResourceList{}
}

func (d *DataSourceRadarResourceList) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_radar_resource_list"
}

func (d *DataSourceRadarResourceList) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}

func (d *DataSourceRadarResourceList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of radar resource data.",
		Attributes: map[string]schema.Attribute{
			// TODO: Figure out how to auto set this like we do elsewhere, cant use plan modifier because this is a data source.
			//"project_id": schema.StringAttribute{
			//	Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
			//	Required:    true,
			//},
			"uri_like_filter": schema.ListAttribute{
				Required:    true,
				Description: "List of uri like filters to apply radar resources",
				ElementType: types.StringType,
			},
			"uri_like_filter_case_insensitive": schema.BoolAttribute{
				Description: "If true, the uri like filter will be case insensitive. Defaults to false.",
				Optional:    true,
			},
			"resources": schema.ListNestedAttribute{
				Description: "TODO",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource id ",
						},
						"name": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource name",
						},
						"uri": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource uri",
						},
						"connection_url": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource connection url",
						},
						"data_source_name": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource data source name",
						},
						"data_source_type": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource data source type",
						},
						"data_source_info": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource data source info",
						},
						"description": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource description",
						},
						"detector_type": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource detector type",
						},
						"visibility": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource visibility",
						},
						"state": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource state",
						},
						"hcp_resource_id": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource HCP resource ID",
						},
						"hcp_resource_name": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource HCP resource name",
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceRadarResourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Validate that the project id and config are set in the config.

	var data DataSourceRadarResourceListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.URILikeFilterCaseInsensitive.IsUnknown() {
		data.URILikeFilterCaseInsensitive = types.BoolValue(false)
	}

	uriLikeFilter := make([]string, 0, len(data.URILikeFilter))
	for _, uri := range data.URILikeFilter {
		if !uri.IsNull() && !uri.IsUnknown() {
			uriLikeFilter = append(uriLikeFilter, uri.ValueString())
		}
	}

	// Do the search for radar resources
	resource, diag := listRadarResources(ctx, d.client, uriLikeFilter, data.URILikeFilterCaseInsensitive.ValueBool())
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	data.Resources = resource
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// listRadarResources retrieves radar resources based on the provided filters.
// It handles the pagination and returns a slice of Resource objects.
func listRadarResources(ctx context.Context, client *clients.Client, uriLikeFilter []string, caseInsensitive bool) ([]Resource, diag.Diagnostics) {
	values := make([]*models.VaultRadar20230501FilterValue, 0, len(uriLikeFilter))
	for _, uri := range uriLikeFilter {
		values = append(values, &models.VaultRadar20230501FilterValue{StringValue: uri})
	}

	// Use LIKE for case-sensitive search, otherwise use ILIKE.
	likeOp := models.FilterFilterOperationLIKE
	if caseInsensitive {
		likeOp = models.FilterFilterOperationILIKE
	}

	// Filter to apply to the radar resources.
	filters := []*models.VaultRadar20230501Filter{
		{
			ID:         "uri",
			Op:         models.NewFilterFilterOperation(likeOp),
			Value:      values,
			ExactMatch: true,
		},
		{
			ID:         "state",
			Op:         models.NewFilterFilterOperation(models.FilterFilterOperationNEQNULLAWARE),
			Value:      []*models.VaultRadar20230501FilterValue{{StringValue: "deleted"}},
			ExactMatch: true,
		},
	}

	var pageLimit = 1000
	resource := make([]Resource, 0, pageLimit) // Note this just initializes the slice with a capacity of pageLimit, but it can grow as needed in the case we have to paginate through the results.

	// Use a loop to paginate through the results.
	for i := 1; ; i++ {
		body := rrs.ListResourcesBody{
			Location: &rrs.ListResourcesParamsBodyLocation{
				OrganizationID: client.Config.OrganizationID,
			},
			Search: &models.VaultRadar20230501SearchSchema{
				Limit:   int32(pageLimit),
				Page:    int32(i),
				Filters: filters,
				// specify order to make the pagination consistent.
				OrderBy:        "uri",
				OrderDirection: "ASC",
			},
		}

		res, err := clients.ListRadarResources(ctx, client, client.Config.ProjectID, body)
		if err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to list radar resources",
					fmt.Sprintf("Error listing radar resources: %s", err.Error()),
				),
			}
		}

		n := len(res.Payload.Resources)
		page := make([]Resource, 0, n)
		for _, resource := range res.Payload.Resources {
			r := toResource(resource)
			page = append(page, r)
		}

		resource = append(resource, page...)

		if n < pageLimit {
			// No more resources to process, break the loop.
			break
		}
	}

	return resource, diag.Diagnostics{}
}

func toResource(resource *models.VaultRadar20230501Resource) Resource {
	r := Resource{
		ID:              basetypes.NewStringValue(resource.ID),
		Name:            basetypes.NewStringValue(resource.Name),
		URI:             basetypes.NewStringValue(resource.URI),
		ConnectionURL:   basetypes.NewStringValue(resource.ConnectionURL),
		DataSourceName:  basetypes.NewStringValue(resource.DataSourceName),
		DataSourceType:  basetypes.NewStringValue(resource.DataSourceType),
		DataSourceInfo:  basetypes.NewStringValue(resource.DataSourceInfo),
		Description:     basetypes.NewStringValue(resource.Description),
		DetectorType:    basetypes.NewStringValue(resource.DetectorType),
		Visibility:      basetypes.NewStringValue(resource.Visibility),
		State:           basetypes.NewStringValue(resource.State),
		HCPResourceID:   basetypes.NewStringValue(resource.HcpResourceID),
		HCPResourceName: basetypes.NewStringValue(resource.HcpResourceName),
	}
	return r
}
