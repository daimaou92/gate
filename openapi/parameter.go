package openapi

import "fmt"

type ParameterObject struct {
	Name            string           `json:"name"` // Required
	In              In               `json:"in"`   // Required
	Description     string           `json:"description"`
	Required        bool             `json:"required"`
	Deprecated      bool             `json:"deprecated"`
	AllowEmptyValue bool             `json:"allowEmptyValue"`
	Style           Style            `json:"style,omitempty"`
	Explode         bool             `json:"explode,omitempty"`
	AllowReserved   bool             `json:"allowReserved,omitempty"`
	Schema          *Schema          `json:"schema,omitempty"`
	Example         interface{}      `json:"example,omitempty"`
	Examples        Examples         `json:"examples,omitempty"` // accepts ExampleObject | ReferenceObject
	Content         *MediaTypeObject `json:"content,omitempty"`
}

func (po ParameterObject) Assert() error {
	if po.Name == "" {
		return wrapErr(fmt.Errorf("\"Name\" is a required field"))
	}

	if err := po.In.Assert(); err != nil {
		return wrapErr(err)
	}

	if err := po.Examples.Assert(); err != nil {
		return wrapErr(err)
	}
	// Match Style with In
	if po.Style != "" {
		if err := po.Style.Assert(); err != nil {
			return wrapErr(err)
		}

		switch po.Style {
		case STYLE_MATRIX:
			if po.In != IN_PATH {
				return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\" for \"Style\": \"%s\"", po.In, po.Style))
			}
		case STYLE_LABEL:
			if po.In != IN_PATH {
				return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\" for \"Style\": \"%s\"", po.In, po.Style))
			}
		case STYLE_FORM:
			if po.In != IN_QUERY && po.In != IN_COOKIE {
				return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\" for \"Style\": \"%s\"", po.In, po.Style))
			}
		case STYLE_SIMPLE:
			if po.In != IN_PATH && po.In != IN_HEADER {
				return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\" for \"Style\": \"%s\"", po.In, po.Style))
			}
		case STYLE_SPACE_DELIMITED, STYLE_PIPE_DELIMITED, STYLE_DEEP_OBJECT:
			if po.In != IN_QUERY {
				return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\" for \"Style\": \"%s\"", po.In, po.Style))
			}
		default:
			return wrapErr(fmt.Errorf("invalid value for \"Style\": \"%s\"", po.Style))
		}
	}
	return nil
}
