package openapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
)

type SpecificationExtension struct {
	ExtensionMap map[string]interface{} `json:"-" yaml:"-"`
}

func (se SpecificationExtension) Extension() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range se.ExtensionMap {
		if !strings.HasPrefix(strings.ToLower(k), "x-") {
			continue
		}
		m[k] = v
	}
	return m
}

type SpecExtended interface {
	Extension() map[string]interface{}
}

// This will generate a map including all the extensions
// matching OAS
// Extensions won't overwrite values
func MapStructSpecExtension(v SpecExtended) (map[string]interface{}, error) {
	e := reflect.TypeOf(v)
	if e.Kind() == reflect.Ptr {
		e = e.Elem()
	}

	if e.Kind() != reflect.Struct {
		return nil, wrapErr(fmt.Errorf("expecting struct or struct pointer"))
	}

	bs, err := json.Marshal(v)
	if err != nil {
		return nil, wrapErr(err, "json marshal failed")
	}

	t := map[string]interface{}{}
	if err := json.Unmarshal(bs, &t); err != nil {
		return nil, wrapErr(err, "unmarshal to map failed")
	}

	for k, v := range v.Extension() {
		if _, ok := t[k]; ok {
			continue
		}
		t[k] = v
	}
	return t, nil
}
