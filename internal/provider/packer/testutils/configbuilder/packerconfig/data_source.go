// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerconfig

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder"
)

func newPackerDataSourceBuilder(sourceType string, uniqueName string) configbuilder.DataSourceBuilder {
	return configbuilder.NewDataSourceBuilder(
		fmt.Sprintf("hcp_packer_%s", sourceType),
		uniqueName,
	)
}
