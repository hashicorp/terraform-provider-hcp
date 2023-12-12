// iam package contains resources and data sources for managing HCP IAM features.
package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/iam/sources/serviceprincipal"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/iam/sources/serviceprincipalkey"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/iam/sources/workloadidentityprovider"
)

// RESOURCE_BUILDERS is a list of all HCP IAM resources exposed by the
// Framework provider. To add a new resource, add a new function to this list.
var RESOURCE_BUILDERS []func() resource.Resource = []func() resource.Resource{
	serviceprincipal.NewResource,
	serviceprincipalkey.NewResource,
	workloadidentityprovider.NewResource,
}

// DATA_SOURCE_BUILDERS is a list of all HCP IAM data sources exposed by the
// Framework provider. To add a new data source, add a new function to this list.
var DATA_SOURCE_BUILDERS []func() datasource.DataSource = []func() datasource.DataSource{
	serviceprincipal.NewDataSource,
}
