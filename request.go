package gate

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

type QueryPayload url.Values

func (qp QueryPayload) Marshal() ([]byte, error) {
	return json.Marshal(qp)
}

func (qp *QueryPayload) Unmarshal(bs []byte) error {
	j := map[string][]string{}
	if err := json.Unmarshal(bs, &j); err != nil {
		return wrapErr(err)
	}
	*qp = QueryPayload(j)
	return nil
}

func (QueryPayload) ContentType() ContentType {
	return ContentTypeJSON
}

type JSONContent struct {
	content interface{}
}

func NewJSONContent(d interface{}) (*JSONContent, error) {
	switch reflect.TypeOf(d).Kind() {
	case reflect.Func, reflect.Chan, reflect.Interface,
		reflect.Ptr, reflect.Uintptr, reflect.UnsafePointer,
		reflect.Invalid:
		return nil, wrapErr(fmt.Errorf("invalid type for JSON: %s", reflect.TypeOf(d).Kind().String()))
	}
	return &JSONContent{
		content: d,
	}, nil
}

func (jc JSONContent) Marshal() ([]byte, error) {
	return json.Marshal(jc.content)
}

func (jc *JSONContent) Unmarshal(bs []byte) error {
	return json.Unmarshal(bs, &jc.content)
}

func (JSONContent) ContentType() ContentType {
	return ContentTypeJSON
}
