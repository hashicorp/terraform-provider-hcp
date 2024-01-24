package testcheck

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder"
)

type resourceAttrStateSourceInterface interface {
	configbuilder.HasBlockName
	configbuilder.HasAttributes
}

func AttributeSet(sourceConfig resourceAttrStateSourceInterface, attrKey string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrSet(sourceConfig.BlockName(), attrKey)
}

func Attribute(sourceConfig resourceAttrStateSourceInterface, attrKey string, value string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttr(sourceConfig.BlockName(), attrKey, value)
}
