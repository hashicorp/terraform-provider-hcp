// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configbuilder

import (
	"fmt"
	"maps"
	"strings"
)

// TODO: Add support for blocks
type Builder interface {
	HasBlockName
	HasBlockIdentifier

	HasBlockLabels
	HasAttributes

	ConfigStringer
}

func BuildersToString(builders ...Builder) string {
	configs := []string{}
	for _, builder := range builders {
		configs = append(configs, builder.ToConfigString())
	}
	return strings.Join(configs, "\n")
}

type HasBlockIdentifier interface {
	// Unquoted Block identifier, like `resource`, `data`, or `output`
	BlockIdentifier() string
}

type HasSourceType interface {
	// Unquoted source type
	//
	// Example: "hcp_packer_version" for `data.hcp_packer_version.prod`
	SourceType() string
}

type sourceTypeMixin struct {
	// Unquoted source type, must not be empty
	sourceType string
}

var _ HasSourceType = sourceTypeMixin{}

func (m sourceTypeMixin) SourceType() string {
	if m.sourceType == "" {
		panic("expected non-empty string for SourceType, but has empty string")
	}

	return m.sourceType
}

func NewSourceTypeMixin(sourceType string) HasSourceType {
	return sourceTypeMixin{
		sourceType: sourceType,
	}
}

type HasUniqueName interface {
	// Unquoted unique name, like for resources/data sources/inputs/outputs
	//
	// Example: "prod" for `data.hcp_packer_version.prod`
	UniqueName() string
}

func NewUniqueNameMixin(uniqueName string) HasUniqueName {
	return uniqueNameMixin{
		uniqueName: uniqueName,
	}
}

type uniqueNameMixin struct {
	// Unquoted unique name, must not be empty
	uniqueName string
}

var _ HasUniqueName = uniqueNameMixin{}

func (m uniqueNameMixin) UniqueName() string {
	if m.uniqueName == "" {
		panic("expected non-empty string for UniqueName, but has empty string")
	}

	return m.uniqueName
}

type HasBlockName interface {
	// The fully-qualified name of a block that can be referenced
	//
	// Example:
	// `hcp_packer_channel.prod` (a resource) or
	// `data.hcp_packer_version.latest` (a data source)
	BlockName() string
}

type SourceBlockNameMixinPrerequisitesInterface interface {
	HasBlockIdentifier
	HasSourceType
	HasUniqueName
}

// NewSourceBlockNameMixin creates a new block name mixin for a resource or
// data source.
func NewSourceBlockNameMixin(blockname SourceBlockNameMixinPrerequisitesInterface) HasBlockName {
	return sourceBlockNameMixin{
		SourceBlockNameMixinPrerequisitesInterface: blockname,
	}
}

type sourceBlockNameMixin struct {
	SourceBlockNameMixinPrerequisitesInterface
}

var _ HasBlockName = sourceBlockNameMixin{}

func (m sourceBlockNameMixin) BlockName() string {
	if m.BlockIdentifier() == "" {
		panic("expected non-empty string for BlockIdentifier, but got empty string")
	}
	if m.SourceType() == "" {
		panic("expected non-empty string for SourceType, but got empty string")
	}
	if m.UniqueName() == "" {
		panic("expected non-empty string for UniqueName, but got empty string")
	}

	return fmt.Sprintf("%s.%s.%s", m.BlockIdentifier(), m.SourceType(), m.UniqueName())
}

type HasBlockLabels interface {
	// Unquoted Block labels
	// The SourceType should be the first entry if HasSourceType is true.
	// The UniqueName should be the second entry if HasUniqueName is true. (first if HasSourceType is false)
	BlockLabels() []string
}

type DefaultBlockLabelsMixinPrerequisitesInterface interface {
	HasSourceType
	HasUniqueName
}

// NewDefaultBlockLabelsMixin creates a new defaultBlockLabelsMixin with the given
// prerequisites. The default block labels are the SourceType and UniqueName, in
// that order, omitted if they are set to empty strings.
func NewDefaultBlockLabelsMixin(prereqs DefaultBlockLabelsMixinPrerequisitesInterface) HasBlockLabels {
	return defaultBlockLabelsMixin{
		DefaultBlockLabelsMixinPrerequisitesInterface: prereqs,
	}
}

type defaultBlockLabelsMixin struct {
	DefaultBlockLabelsMixinPrerequisitesInterface
}

var _ HasBlockLabels = defaultBlockLabelsMixin{}

func (m defaultBlockLabelsMixin) BlockLabels() []string {
	return []string{m.SourceType(), m.UniqueName()}
}

type HasAttributes interface {
	// Attributes for the block, keyed by attribute name
	// The value should be the literal text value of the attribute,
	// as it would be typed in a terraform configuration file.
	// May contain terraform expressions and attribute references.
	//
	// Examples:
	// greeting: `"hello"` for `greeting = "hello"`
	// version_fingerprint: `data.hcp_packer_version.latest.fingerprint` for `version_fingerprint = data.hcp_packer_version.latest.fingerprint`
	Attributes() map[string]string

	GetAttribute(string) string
	SetAttribute(string, string)
}

func NewAttributesMixin(initialAttributes map[string]string) HasAttributes {
	if initialAttributes == nil {
		initialAttributes = map[string]string{}
	}
	return &attributesMixin{
		attributes: maps.Clone(initialAttributes),
	}
}

type attributesMixin struct {
	// Attributes for the block
	attributes map[string]string
}

var _ HasAttributes = &attributesMixin{}

func (m attributesMixin) Attributes() map[string]string {
	return maps.Clone(m.attributes)
}

func (m *attributesMixin) GetAttribute(attributeName string) string {
	value, ok := m.attributes[attributeName]
	if !ok {
		return ""
	}

	return value
}

func (m *attributesMixin) SetAttribute(attributeName string, value string) {
	m.attributes[attributeName] = value
}

type CanReferenceAttributes interface {
	AttributeRefMixinInterface
	AttributeRefStrictMixinInterface
}

type AttributeRefMixinInterface interface {
	// Should return `BlockName()+"."+attributeName` for blocks that can be referenced.
	// Resources may or may not contain have the `resource.` suffix in the reference.
	AttributeRef(string) string
}

func NewAttributeRefMixin(blockName HasBlockName) AttributeRefMixinInterface {
	return attributeRefMixin{
		HasBlockName: blockName,
	}
}

type attributeRefMixin struct {
	HasBlockName
}

var _ AttributeRefMixinInterface = attributeRefMixin{}

func (m attributeRefMixin) AttributeRef(attributeName string) string {
	return fmt.Sprintf("%s.%s", m.BlockName(), attributeName)
}

type AttributeRefStrictMixinInterface interface {
	// Should return `BlockName()+"."+attributeName` for blocks that can be referenced.
	// Resources may or may not contain have the `resource.` suffix in the reference.
	// Panics if the block cannot be referenced (ex. output blocks)
	// Panics if the attribute is not present in Attributes()
	AttributeRefStrict(string) string
}

type AttributesRefStrictMixinPrerequisitesInterface interface {
	HasAttributes
	AttributeRefMixinInterface
}

func NewAttributesRefStrictMixin(prereqs AttributesRefStrictMixinPrerequisitesInterface) AttributeRefStrictMixinInterface {
	return attributesRefStrictMixin{
		AttributesRefStrictMixinPrerequisitesInterface: prereqs,
	}
}

type attributesRefStrictMixin struct {
	AttributesRefStrictMixinPrerequisitesInterface
}

var _ AttributeRefStrictMixinInterface = attributesRefStrictMixin{}

func (m attributesRefStrictMixin) AttributeRefStrict(attributeName string) string {
	if _, ok := m.Attributes()[attributeName]; !ok {
		panic(fmt.Errorf("attribute %q not found in Attributes()", attributeName))
	}

	return m.AttributeRef(attributeName)
}

type ConfigStringer interface {
	// Returns the HCL multiline configuration string for the block
	ToConfigString() string
}

type defaultConfigStringerMixinPrerequisitesInterface interface {
	HasBlockIdentifier
	HasBlockLabels
	HasAttributes
}

// DefaultConfigStringerMixinPrerequisites can be used to merge component elements of the
// default config stringer mixin prerequisites interface into a single object if there
// is not another way to meet the interface requirements.
type DefaultConfigStringerMixinPrerequisites struct {
	HasBlockIdentifier
	HasBlockLabels
	HasAttributes
}

var _ defaultConfigStringerMixinPrerequisitesInterface = DefaultConfigStringerMixinPrerequisites{}

// NewDefaultConfigStringerMixin creates a new defaultConfigStringerMixin with the given prerequisites.
func NewDefaultConfigStringerMixin(prereqs defaultConfigStringerMixinPrerequisitesInterface) ConfigStringer {
	return defaultConfigStringerMixin{
		defaultConfigStringerMixinPrerequisitesInterface: prereqs,
	}
}

type defaultConfigStringerMixin struct {
	defaultConfigStringerMixinPrerequisitesInterface
}

var _ ConfigStringer = defaultConfigStringerMixin{}

func (m defaultConfigStringerMixin) ToConfigString() string {
	labelsString := ""
	for _, label := range m.BlockLabels() {
		labelsString += fmt.Sprintf("%q ", label)
	}

	attributesString := ""
	for key, value := range m.Attributes() {
		if key != "" && value != "" {
			attributesString += fmt.Sprintf("	%s = %s\n", key, value)
		}
	}

	return fmt.Sprintf(`
	%s %s{
	%s}
	`,
		m.BlockIdentifier(), labelsString,
		attributesString,
	)

}
