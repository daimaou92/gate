package openapi

import (
	"reflect"

	"github.com/goccy/go-json"
)

type Schema struct {
	Title                string                `json:"title,omitempty" yaml:"title,omitempty"`
	MultipleOf           *float64              `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              *float64              `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     bool                  `json:"exclusiveMaximum,omitempty" yaml:"exclusivemaximum,omitempty"`
	Minimum              *float64              `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     bool                  `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            uint                  `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            uint                  `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string                `json:"pattern,omitempty" yaml:"pattern,omitempty"` // should be a valid regex when exists
	MaxItems             uint                  `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             uint                  `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool                  `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        uint                  `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        uint                  `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required             []string              `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}         `json:"enum,omitempty" yaml:"enum,omitempty"`
	Type                 Type                  `json:"type,omitempty" yaml:"type,omitempty"`
	AllOf                []*SchemaRef          `json:"allOf,omitempty" yaml:"allOf,omitempty"` // array of either schema object or ref object
	OneOf                []*SchemaRef          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*SchemaRef          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *SchemaRef            `json:"not,omitempty" yaml:"not,omitempty"`
	Items                *SchemaRef            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*SchemaRef `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties interface{}           `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"` // SchemaRef or bool. Default = true
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Format               Format                `json:"format,omitempty" yaml:"format,omitempty"`
	Default              interface{}           `json:"default,omitempty" yaml:"default,omitempty"`
	Nullable             bool                  `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	DiscriminatorObject  *DiscriminatorObject  `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	ReadOnly             bool                  `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                  `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	XML                  *XMLObject            `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Example              interface{}           `json:"example,omitempty" yaml:"example,omitempty"`
	Deprecated           bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	SpecificationExtension
}

type SchemaRef struct {
	Schema *Schema
	Ref    *ReferenceObject
}

func (sr SchemaRef) determineValue() interface{} {
	if sr.Schema != nil {
		return sr.Schema
	}

	if sr.Ref != nil {
		return sr.Ref
	}
	return nil
}

func (sr SchemaRef) MarshalJSON() ([]byte, error) {
	bs, err := json.Marshal(sr.determineValue())
	if err != nil {
		return nil, wrapErr(err, "json marshal failed")
	}
	return bs, nil
}

func (sr *SchemaRef) UnmarshalJSON(src []byte) error {
	var s Schema
	if err := json.Unmarshal(src, &s); err == nil {
		if !reflect.ValueOf(s).IsZero() {
			if sr == nil {
				sr = &SchemaRef{}
			}
			sr.Schema = &s
			return nil
		}
	}

	var r ReferenceObject
	if err := json.Unmarshal(src, &r); err == nil {
		if !reflect.ValueOf(r).IsZero() {
			if sr == nil {
				sr = &SchemaRef{}
			}
			sr.Ref = &r
		}
	} else {
		return wrapErr(err, "coundn't json unmarshal into Schema or ReferenceObject")
	}
	return nil
}
