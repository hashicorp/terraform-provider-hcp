package vaultsecrets

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type GatewayPool struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`

	//ResourceName   types.String `tfsdk:"resource_name"`
	//ResourceID     types.String `tfsdk:"resource_id"`

	// other fields in the response
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	CertPem      types.String `tfsdk:"cert_pem"`
}

var _ resource.Resource = &resourceVaultSecretsGatewayPool{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsGatewayPool{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsGatewayPool{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsGatewayPool{}

func NewVaultSecretsGatewayPoolResource() resource.Resource {
	return &resourceVaultSecretsGatewayPool{}
}

type resourceVaultSecretsGatewayPool struct {
	client *clients.Client
}

func (r *resourceVaultSecretsGatewayPool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_gateway_pool"
}

func (r *resourceVaultSecretsGatewayPool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: `Name of the gateway pool`,
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: `Description of the gateway pool`,
			Optional:    true,
		},
		//"resource_name": schema.StringAttribute{
		//	Computed: true,
		//},
		//"resource_id": schema.StringAttribute{
		//	Computed: true,
		//},
		"client_id": schema.StringAttribute{
			Computed: true,
		},
		"client_secret": schema.StringAttribute{
			Computed:  true,
			Sensitive: true,
		},
		"cert_pem": schema.StringAttribute{
			Computed:  true,
			Sensitive: true,
		},
	}

	maps.Copy(attributes, locationAttributes)

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *resourceVaultSecretsGatewayPool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*GatewayPool](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		gatewayPool, ok := i.(*GatewayPool)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *GatewayPool, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.CreateGatewayPool(
			secret_service.NewCreateGatewayPoolParamsWithContext(ctx).
				WithOrganizationID(gatewayPool.OrganizationID.ValueString()).
				WithProjectID(gatewayPool.ProjectID.ValueString()).
				WithBody(&secretmodels.SecretServiceCreateGatewayPoolBody{
					Description: gatewayPool.Description.ValueString(),
					Name:        gatewayPool.Name.ValueString(),
				}), nil)

		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}

		// NOTE: we do not return the model here like other `hvsResource`'s
		return response.Payload, nil
	})...)
}

func (r *resourceVaultSecretsGatewayPool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*GatewayPool](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		gatewayPool, ok := i.(*GatewayPool)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *GatewayPool, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetGatewayPool(
			secret_service.NewGetGatewayPoolParamsWithContext(ctx).
				WithOrganizationID(gatewayPool.OrganizationID.ValueString()).
				WithProjectID(gatewayPool.ProjectID.ValueString()).
				WithGatewayPoolName(gatewayPool.Name.ValueString()),
			nil)
		if err != nil && !clients.IsResponseForbidden(err) { // The HVS API returns 403 if the app doesn't exist even if the principal has the correct permissions.
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}

		return response.Payload.GatewayPool, nil
	})...)
}

func (r *resourceVaultSecretsGatewayPool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*GatewayPool](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		gatewayPool, ok := i.(*GatewayPool)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *GatewayPool, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.UpdateGatewayPool(
			secret_service.NewUpdateGatewayPoolParamsWithContext(ctx).
				WithOrganizationID(gatewayPool.OrganizationID.ValueString()).
				WithProjectID(gatewayPool.ProjectID.ValueString()).
				WithGatewayPoolName(gatewayPool.Name.ValueString()).
				WithBody(&secretmodels.SecretServiceUpdateGatewayPoolBody{
					Description: gatewayPool.Description.ValueString(),
				}),
			nil)

		if err != nil && !clients.IsResponseForbidden(err) { // The HVS API returns 403 if the app doesn't exist even if the principal has the correct permissions.
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}

		return response.Payload.GatewayPool, nil
	})...)
}

func (r *resourceVaultSecretsGatewayPool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*GatewayPool](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		gatewayPool, ok := i.(*GatewayPool)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *GatewayPool, got %T", i)
		}

		_, err := r.client.VaultSecrets.DeleteGatewayPool(
			secret_service.NewDeleteGatewayPoolParamsWithContext(ctx).
				WithOrganizationID(gatewayPool.OrganizationID.ValueString()).
				WithProjectID(gatewayPool.ProjectID.ValueString()).
				WithGatewayPoolName(gatewayPool.Name.ValueString()),
			nil)

		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsGatewayPool) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsGatewayPool) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsGatewayPool) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_name"), req.ID)...)
}

var _ hvsResource = &GatewayPool{}

func (g *GatewayPool) projectID() types.String {
	return g.ProjectID
}

func (g *GatewayPool) initModel(_ context.Context, orgID, projID string) diag.Diagnostics {
	g.OrganizationID = types.StringValue(orgID)
	g.ProjectID = types.StringValue(projID)

	return diag.Diagnostics{}
}

func (g *GatewayPool) fromModel(_ context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	g.OrganizationID = types.StringValue(orgID)
	g.ProjectID = types.StringValue(projID)

	switch v := model.(type) {
	case *secretmodels.Secrets20231128CreateGatewayPoolResponse:
		// create pool gives different values that we want
		g.Name = types.StringValue(v.GatewayPool.Name)
		g.Description = types.StringValue(v.GatewayPool.Description)
		g.CertPem = types.StringValue(v.CertPem)
		g.ClientID = types.StringValue(v.ClientID)
		g.ClientSecret = types.StringValue(v.ClientSecret)
	case *secretmodels.Secrets20231128GatewayPool:
		g.Name = types.StringValue(v.Name)
		g.Description = types.StringValue(v.Description)
	default:
		diags.AddError("fromModel", fmt.Sprintf("invalid model type: %T", model))
	}

	return diags
}
