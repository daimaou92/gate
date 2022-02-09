package openapi

import "fmt"

type HeaderObject struct{}

type Headers map[string]interface{}

func (h Headers) Assert() error {
	if h == nil {
		return nil
	}

	for k, v := range h {
		if _, ok := v.(HeaderObject); !ok {
			if _, ok := v.(ReferenceObject); !ok {
				return wrapErr(fmt.Errorf("value for key: \"%s\" is neither a ReferenceObject nor a HeaderObject", k))
			}
		}
	}
	return nil
}
