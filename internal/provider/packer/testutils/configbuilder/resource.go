package configbuilder

import "fmt"

type ResourceBuilder interface {
	Builder
	HasSourceType
	HasUniqueName
	CanReferenceAttributes

	// The fully-qualified name of a resource that can be referenced
	//
	// Should be formatted like `hcp_packer_channel.prod`, similar to
	// BlockName(), but without the "resource." prefix
	ResourceName() string
}

func NewResourceBuilder(sourceType string, uniqueName string) ResourceBuilder {
	return newResourceBuilder(sourceType, uniqueName, nil)
}

func CloneResourceBuilder(oldBuilder ResourceBuilder) ResourceBuilder {
	return newResourceBuilder(oldBuilder.SourceType(), oldBuilder.UniqueName(), oldBuilder.Attributes())
}

func newResourceBuilder(sourceType string, uniqueName string, initialAttributes map[string]string) ResourceBuilder {
	resource := &resourceBuilder{}

	resource.HasSourceType = NewSourceTypeMixin(sourceType)
	resource.HasUniqueName = NewUniqueNameMixin(uniqueName)

	resource.HasBlockName = NewSourceBlockNameMixin(resource)
	resource.HasBlockLabels = NewDefaultBlockLabelsMixin(resource)

	resource.HasAttributes = NewAttributesMixin(initialAttributes)
	resource.AttributeRefStrictMixinInterface = NewAttributesRefStrictMixin(resource)

	resource.ConfigStringer = NewDefaultConfigStringerMixin(resource)

	return resource
}

type resourceBuilder struct {
	HasSourceType
	HasUniqueName

	HasBlockName
	HasBlockLabels

	HasAttributes
	// AttributeRef has a custom implementation for resources
	AttributeRefStrictMixinInterface

	ConfigStringer
}

var _ ResourceBuilder = &resourceBuilder{}

func (b resourceBuilder) BlockIdentifier() string {
	return "resource"
}

func (b resourceBuilder) ResourceName() string {
	return b.BlockName()
}

// Custom AttributeRef implementation for resources to hide the "resource." prefix
// from the beginning of the resource name
func (b resourceBuilder) AttributeRef(attributeName string) string {
	return fmt.Sprintf("%s.%s", b.ResourceName(), attributeName)
}
