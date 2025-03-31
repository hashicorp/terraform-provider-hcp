// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceUserPrincipals struct {
	client *clients.Client
}

type UserPrincipalModel struct {
	UserID types.String `tfsdk:"user_id"`
	Email  types.String `tfsdk:"email"`
}

type DataSourceUserPrincipalsModel struct {
	Users []UserPrincipalModel `tfsdk:"users"`
}

func NewUserPrincipalsDataSource() datasource.DataSource {
	return &DataSourceUserPrincipals{}
}

func (d *DataSourceUserPrincipals) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_principals"
}

func (d *DataSourceUserPrincipals) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The user principals data source retrieves all user principals in the organization.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "List of user principals in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "The user's unique identifier.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The user's email.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceUserPrincipals) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceUserPrincipals) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceUserPrincipalsModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	// Search for all user principals in the organization
	getParams := iam_service.NewIamServiceSearchPrincipalsParamsWithContext(ctx)
	getParams.SetOrganizationID(d.client.Config.OrganizationID)
	getParams.SetBody(iam_service.IamServiceSearchPrincipalsBody{
		Filter: &models.HashicorpCloudIamSearchPrincipalsFilter{
			// Empty search text will return all principals
			SearchText: "",
		},
	})

	res, err := d.client.IAM.IamServiceSearchPrincipals(getParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving user principals", err.Error())
		return
	}

	// Convert the response to our model
	users := make([]UserPrincipalModel, 0)
	for _, principal := range res.Payload.Principals {
		// Only include user principals (not service principals)
		if principal.Email != "" {
			users = append(users, UserPrincipalModel{
				UserID: types.StringValue(principal.ID),
				Email:  types.StringValue(principal.Email),
			})
		}
	}

	data.Users = users
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
