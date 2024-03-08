// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceUserPrincipal struct {
	client *clients.Client
}

type DataSourceUserPrincipalModel struct {
	UserID types.String `tfsdk:"user_id"`
	Email  types.String `tfsdk:"email"`
}

func NewUserPrincipalDataSource() datasource.DataSource {
	return &DataSourceUserPrincipal{}
}

func (d *DataSourceUserPrincipal) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_principal"
}

func (d *DataSourceUserPrincipal) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The user principal data source retrieves the given user principal.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Description: "The user's unique identifier. Can not be combined with email.",
				Computed:    true,
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The user's email. Can not be combined with user_id.",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

func (d *DataSourceUserPrincipal) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceUserPrincipal) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceUserPrincipalModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}
	if data.UserID.IsNull() && data.Email.IsNull() {
		// Both user_id and email were not provided which is not allowed.
		resp.Diagnostics.AddError(
			"Invalid input",
			"Either user_id or email must be set in your input.",
		)
	} else if !data.UserID.IsNull() && !data.Email.IsNull() {
		// Both user_id and email were provided which is not allowed.
		resp.Diagnostics.AddError(
			"Invalid input",
			"Both email and user_id can not be set at the same time. Either use only email or user_id in your input and try again.",
		)
		return
	} else if !data.UserID.IsNull() {
		// Get the user principal by ID.
		getParams := iam_service.NewIamServiceGetUserPrincipalByIDInOrganizationParams()
		getParams.UserPrincipalID = data.UserID.ValueString()

		res, err := d.client.IAM.IamServiceGetUserPrincipalByIDInOrganization(getParams, nil)
		if err != nil {
			var getErr *iam_service.IamServiceGetUserPrincipalByIDInOrganizationDefault
			if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
				resp.Diagnostics.AddError("User principal does not exist", fmt.Sprintf("unknown user principal with ID %q", data.UserID.ValueString()))
				return
			}

			resp.Diagnostics.AddError("Error retrieving user principal", err.Error())
			return
		}

		// Set the user principal data.
		data.UserID = types.StringValue(res.Payload.UserPrincipal.ID)
		data.Email = types.StringValue(res.Payload.UserPrincipal.Email)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	} else if !data.Email.IsNull() {
		// Search for the user principal by email.
		getParams := iam_service.NewIamServiceSearchPrincipalsParams()
		getParams.SetBody(iam_service.IamServiceSearchPrincipalsBody{
			Filter: &models.HashicorpCloudIamSearchPrincipalsFilter{
				SearchText: data.Email.ValueString(),
			},
		})

		res, err := d.client.IAM.IamServiceSearchPrincipals(getParams, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving user principal", err.Error())
			return
		}

		// No user principal found.
		if len(res.Payload.Principals) == 0 {
			resp.Diagnostics.AddError(
				"User principal does not exist",
				fmt.Sprintf("unknown user principal with email %q", data.Email.ValueString()),
			)
			return
		}

		// More than 1 user principal found.
		if len(res.Payload.Principals) > 1 {
			resp.Diagnostics.AddError(
				"Multiple User Principals Found",
				fmt.Sprintf("More than 1 user was found with the specified email address (%q). Please visit the HCP Portal to retrieve the desired user ID.", data.Email.ValueString()),
			)
			return
		}

		// Default to returning the first user principal found.
		data.UserID = types.StringValue(res.Payload.Principals[0].ID)
		data.Email = types.StringValue(res.Payload.Principals[0].Email)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}
