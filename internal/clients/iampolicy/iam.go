package iampolicy

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// The ResourceIamUpdater interface is implemented for each HCP resource
// supporting IAM policy (Organization/Project/Resource).
//
// Implementations should be created per resource and should keep track of the
// resource identifier.
type ResourceIamUpdater interface {
	// Fetch the existing IAM policy attached to a resource.
	GetResourceIamPolicy(context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics)

	// Replaces the existing IAM Policy attached to a resource.
	SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics)

	// A mutex guards against concurrent to call to the SetResourceIamPolicy method.
	// The mutex key should be globally unique.
	GetMutexKey() string

	// Returns the unique resource identifier.
	//GetResourceId() string

	// Textual description of this resource to be used in error message.
	// The description should include the unique resource identifier.
	//DescribeResource() string
}

type TerraformResourceData interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
	SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics
}

// Factory for generating ResourceIamUpdater for given ResourceData resource
type NewResourceIamUpdaterFunc func(ctx context.Context, d TerraformResourceData, clients *clients.Client) (ResourceIamUpdater, diag.Diagnostics)
