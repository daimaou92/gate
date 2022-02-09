package openapi

type XMLObject struct {
	SpecificationExtension

	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace:omitempty" yaml:"namespace:omitempty"`
	Prefix    string `json:"prefix:omitempty" yaml:"prefix:omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"attribute,omitempty"`
}
