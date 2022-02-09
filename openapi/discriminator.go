package openapi

import "fmt"

type DiscriminatorObject struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

func (do DiscriminatorObject) Assert() error {
	if do.PropertyName == "" {
		return wrapErr(fmt.Errorf("propertyName cannot be empty"))
	}
	return nil
}
