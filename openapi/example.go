package openapi

import "fmt"

type ExampleObject struct {
	Summary       string      `json:"summary"`
	Description   string      `json:"description"`
	Value         interface{} `json:"value"`
	ExternalValue string      `json:"externalValue"`
}

type Examples map[string]interface{}

func (e Examples) Assert() error {
	if e == nil {
		return nil
	}
	for k, v := range e {
		if _, ok := v.(ExampleObject); !ok {
			if _, ok := v.(ReferenceObject); !ok {
				return wrapErr(fmt.Errorf("value for key: \"%s\" is neither a ReferenceObject nor an ExampleObject", k))
			}
		}
	}
	return nil
}
