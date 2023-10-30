package resourcemanager

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/iampolicy"
)

var (
	orgIAMSchema = schema.Schema{
		MarkdownDescription: "Allows definitively setting the organization's IAM policy. " +
			"This will replace any existing policy already attached.",
	}

	_ iampolicy.NewResourceIamUpdaterFunc = newOrgIAMPolicyUpdater
	_ iampolicy.ResourceIamUpdater        = &orgIAMPolicyUpdater{}
)

func NewOrganizationIAMPolicyResource() resource.Resource {
	return iampolicy.NewResourceIamPolicy("organization", orgIAMSchema, "", newOrgIAMPolicyUpdater)
}

func NewOrganizationIAMBindingResource() resource.Resource {
	return iampolicy.NewResourceIamBinding("organization", orgIAMSchema, "", newOrgIAMPolicyUpdater)
}

type orgIAMPolicyUpdater struct {
	client *clients.Client
	d      iampolicy.TerraformResourceData
}

func newOrgIAMPolicyUpdater(
	ctx context.Context,
	d iampolicy.TerraformResourceData,
	clients *clients.Client) (iampolicy.ResourceIamUpdater, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &orgIAMPolicyUpdater{
		client: clients,
		d:      d,
	}, diags
}

func (u *orgIAMPolicyUpdater) GetMutexKey() string {
	return u.client.Config.OrganizationID
}

// Fetch the existing IAM policy attached to a resource.
func (u *orgIAMPolicyUpdater) GetResourceIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := organization_service.NewOrganizationServiceGetIamPolicyParams()
	params.ID = u.client.Config.OrganizationID
	res, err := u.client.Organization.OrganizationServiceGetIamPolicy(params, nil)
	if err != nil {
		diags.AddError("failed to retrieve organization IAM policy", err.Error())
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}

// Replaces the existing IAM Policy attached to a resource.
func (u *orgIAMPolicyUpdater) SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := organization_service.NewOrganizationServiceSetIamPolicyParams()
	params.ID = u.client.Config.OrganizationID
	params.Body = organization_service.OrganizationServiceSetIamPolicyBody{
		Policy: policy,
	}

	res, err := u.client.Organization.OrganizationServiceSetIamPolicy(params, nil)
	if err != nil {
		diags.AddError("failed to retrieve organization IAM policy", err.Error())
		return nil, diags
	}

	return res.GetPayload().Policy, diags
}
