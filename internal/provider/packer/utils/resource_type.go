// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
)

const PackerResourceTypeMidfix string = "packer"

func PackerResourceType(typeSuffix string) string {
	return fmt.Sprintf("%s_%s", PackerResourceTypeMidfix, typeSuffix)
}
