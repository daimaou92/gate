package openapi

import "fmt"

type ExternalDocs struct {
	SpecificationExtension
	Description string `json:"description"`
	URL         string `json:"url"`
}

func (ed ExternalDocs) Assert() error {
	if ed.URL == "" {
		return wrapErr(fmt.Errorf("URL is a required field"))
	}
	return nil
}
