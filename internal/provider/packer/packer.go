package packer

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/sources/artifact"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/sources/version"
)

// ResourceSchemaBuilders is a list of all HCP Packer resources exposed by the
// Framework provider. To add a new resource, add a new function to this list.
var ResourceSchemaBuilders []func() resource.Resource = []func() resource.Resource{}

// DataSourceSchemaBuilders is a list of all HCP Packer data sources exposed by the
// Framework provider. To add a new data source, add a new function to this list.
var DataSourceSchemaBuilders []func() datasource.DataSource = []func() datasource.DataSource{
	version.NewDataSource,
	artifact.NewDataSource,
}
