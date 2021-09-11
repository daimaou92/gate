package gate

import (
	"encoding/json"
	"log"
	"reflect"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
)

// type Info openapi3.Info

func pathParams(r string) []string {
	reg := regexp.MustCompile(`/:[\p{L}_][\p{L}_\p{Nd}]*`)
	return reg.FindAllString(r, -1)
}

func queryParams(p Payload) []string {
	if p == nil {
		return nil
	}
	bs, err := p.Marshal()
	if err != nil {
		log.Println(wrapErr(err, "Marshal failed"))
		return nil
	}

	v := map[string]interface{}{}
	if err := json.Unmarshal(bs, &v); err != nil {
		log.Println(wrapErr(err, "json unmarshal to map failed"))
		return nil
	}
	var keys []string
	for key := range v {
		keys = append(keys, key)
	}
	return keys
}

func schemaFromType(typ reflect.Type) (openapi3.Schema, error) {
	// TODO
	return *openapi3.NewSchema(), nil
}
