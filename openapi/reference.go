package openapi

import "fmt"

type ReferenceObject struct {
	Ref         string `json:"$ref"` // Required
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func (ro ReferenceObject) Assert() error {
	if ro.Ref == "" {
		return wrapErr(fmt.Errorf("\"Ref\" is required field"))
	}
	return nil
}
