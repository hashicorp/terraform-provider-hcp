package configbuilder

type DataSourceBuilder interface {
	Builder
	HasSourceType
	HasUniqueName
	CanReferenceAttributes

	// DataSourceName is the fully-qualified name of a data source that can be referenced
	// Should be formatted like `data.hcp_packer_version`, the same as BlockName()
	DataSourceName() string
}

func NewDataSourceBuilder(sourceType string, uniqueName string) DataSourceBuilder {
	return newDataSourceBuilder(sourceType, uniqueName, nil)
}

func CloneDataSourceBuilder(oldBuilder DataSourceBuilder) DataSourceBuilder {
	return newDataSourceBuilder(oldBuilder.SourceType(), oldBuilder.UniqueName(), oldBuilder.Attributes())
}

func newDataSourceBuilder(sourceType string, uniqueName string, initialAttributes map[string]string) DataSourceBuilder {
	builder := &dataSourceBuilder{}

	builder.HasSourceType = NewSourceTypeMixin(sourceType)
	builder.HasUniqueName = NewUniqueNameMixin(uniqueName)

	builder.HasBlockName = NewSourceBlockNameMixin(builder)
	builder.HasBlockLabels = NewDefaultBlockLabelsMixin(builder)

	builder.HasAttributes = NewAttributesMixin(initialAttributes)
	builder.AttributeRefMixinInterface = NewAttributeRefMixin(builder)
	builder.AttributeRefStrictMixinInterface = NewAttributesRefStrictMixin(builder)

	builder.ConfigStringer = NewDefaultConfigStringerMixin(builder)

	return builder
}

type dataSourceBuilder struct {
	HasSourceType
	HasUniqueName

	HasBlockName
	HasBlockLabels

	HasAttributes
	AttributeRefMixinInterface
	AttributeRefStrictMixinInterface

	ConfigStringer
}

var _ DataSourceBuilder = &dataSourceBuilder{}

func (b dataSourceBuilder) BlockIdentifier() string {
	return "data"
}

func (b dataSourceBuilder) DataSourceName() string {
	return b.BlockName()
}
