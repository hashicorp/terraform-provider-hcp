// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

type DataSourceRadarResources struct {
	client *clients.Client
}

type DataSourceRadarResourcesModel struct {
	ProjectID     types.String          `tfsdk:"project_id"`
	URILikeFilter ResourceURILikeFilter `tfsdk:"uri_like_filter"`
	Resources     []Resource            `tfsdk:"resources"`
}

type ResourceURILikeFilter struct {
	Values          []types.String `tfsdk:"values"`
	CaseInsensitive types.Bool     `tfsdk:"case_insensitive"`
}

// Resource represents a radar resource in the list of resources.
type Resource struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	URI               types.String `tfsdk:"uri"`
	ConnectionURL     types.String `tfsdk:"connection_url"`
	DataSourceName    types.String `tfsdk:"data_source_name"`
	DataSourceType    types.String `tfsdk:"data_source_type"`
	Description       types.String `tfsdk:"description"`
	DetectorType      types.String `tfsdk:"detector_type"`
	Visibility        types.String `tfsdk:"visibility"`
	State             types.String `tfsdk:"state"`
	HCPResourceName   types.String `tfsdk:"hcp_resource_name"`
	HCPResourceStatus types.String `tfsdk:"hcp_resource_status"`
}

func NewRadarResourcesDataSource() datasource.DataSource {
	return &DataSourceRadarResources{}
}

func (d *DataSourceRadarResources) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_radar_resources"
}

func (d *DataSourceRadarResources) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceRadarResources) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of radar resource data.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
				Optional:    true,
				Computed:    true,
			},
			"uri_like_filter": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Applies a filter to the radar resources based on their URIs. The filter uses the SQL LIKE operator, which allows for wildcard matching.",
				Attributes: map[string]schema.Attribute{
					"values": schema.ListAttribute{
						Required:    true,
						Description: "URI like filters to apply radar resources. Each entry in the list will act like an or condition.",
						ElementType: types.StringType,
					},
					"case_insensitive": schema.BoolAttribute{
						Description: "If true, the uri like filter will be case insensitive. Defaults to false.",
						Optional:    true,
					},
				},
			},
			"resources": schema.ListNestedAttribute{
				Description: "List of Radar resources.",
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
						"hcp_resource_name": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource HCP resource name",
						},
						"hcp_resource_status": &schema.StringAttribute{
							Computed:    true,
							Description: "Radar resource HCP resource status",
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceRadarResources) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceRadarResourcesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := d.client.Config.ProjectID
	if !data.ProjectID.IsNull() {
		projectID = data.ProjectID.ValueString()
	}
	data.ProjectID = types.StringValue(projectID)

	if data.URILikeFilter.CaseInsensitive.IsUnknown() {
		data.URILikeFilter.CaseInsensitive = types.BoolValue(false)
	}

	uriLikeFilter := make([]string, 0, len(data.URILikeFilter.Values))
	for _, uri := range data.URILikeFilter.Values {
		if !uri.IsNull() && !uri.IsUnknown() {
			uriLikeFilter = append(uriLikeFilter, uri.ValueString())
		}
	}

	// Do the search for radar resources
	resource, diag := listRadarResources(ctx, d.client, projectID, uriLikeFilter, data.URILikeFilter.CaseInsensitive.ValueBool())
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the resources in the data model.
	data.Resources = resource

	// Save the data into the Terraform response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// listRadarResources retrieves radar resources based on the provided filters.
// It handles the pagination and returns a slice of Resource objects.
func listRadarResources(ctx context.Context, client *clients.Client, projectID string, uriLikeFilter []string, caseInsensitive bool) ([]Resource, diag.Diagnostics) {
	values := make([]interface{}, 0, len(uriLikeFilter))
	for _, uri := range uriLikeFilter {
		values = append(values, uri)
	}

	// Use LIKE for case-sensitive search, otherwise use ILIKE.
	likeOp := models.VaultRadar20230501FilterV2OperationLIKE
	if caseInsensitive {
		likeOp = models.VaultRadar20230501FilterV2OperationILIKE
	}

	// Filter to apply to the radar resources.
	filters := []*models.VaultRadar20230501FilterV2{
		{
			ID:         "uri",
			Op:         models.NewVaultRadar20230501FilterV2Operation(likeOp),
			Value:      values,
			ExactMatch: true,
		},
		{
			ID:         "state",
			Op:         models.NewVaultRadar20230501FilterV2Operation(models.VaultRadar20230501FilterV2OperationNEQNULLAWARE),
			Value:      []interface{}{"deleted"},
			ExactMatch: true,
		},
	}

	var pageLimit = 1000
	resource := make([]Resource, 0, pageLimit) // Note this just initializes the slice with a capacity of pageLimit, but it can grow as needed in the case we have to paginate through the results.

	// Use a loop to paginate through the results.
	for i := 1; ; i++ {
		body := rrs.SearchResourcesBody{
			Location: &rrs.SearchResourcesParamsBodyLocation{
				OrganizationID: client.Config.OrganizationID,
			},
			Search: &models.VaultRadar20230501SearchSchemaV2{
				Limit:   int32(pageLimit),
				Page:    int32(i),
				Filters: filters,
				// specify order to make the pagination consistent.
				OrderBy:        "uri",
				OrderDirection: "ASC",
			},
		}

		res, err := clients.SearchRadarResources(ctx, client, projectID, body)
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

func toResource(resource *models.VaultRadar20230501ResourceV2) Resource {
	r := Resource{
		ID:                basetypes.NewStringValue(resource.ID),
		Name:              basetypes.NewStringValue(resource.Name),
		URI:               basetypes.NewStringValue(resource.URI),
		ConnectionURL:     basetypes.NewStringValue(resource.ConnectionURL),
		DataSourceName:    basetypes.NewStringValue(resource.DataSourceName),
		DataSourceType:    basetypes.NewStringValue(resource.DataSourceType),
		Description:       basetypes.NewStringValue(resource.Description),
		DetectorType:      basetypes.NewStringValue(resource.DetectorType),
		Visibility:        basetypes.NewStringValue(resource.Visibility),
		State:             basetypes.NewStringValue(resource.State),
		HCPResourceName:   basetypes.NewStringValue(resource.HcpResourceName),
		HCPResourceStatus: basetypes.NewStringValue(resource.HcpResourceStatus),
	}
	return r
}
