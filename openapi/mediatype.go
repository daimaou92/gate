package openapi

type MediaTypeObject struct {
	Schema   *Schema         `json:"schema"`
	Example  interface{}     `json:"example"`
	Examples Examples        `json:"examples"`
	Encoding *EncodingObject `json:"encoding"`
}
