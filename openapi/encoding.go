package openapi

type EncodingObject struct {
	ContentType   string  `json:"contentType"`
	Headers       Headers `json:"headers"`
	Style         Style   `json:"style"`
	Explode       bool    `json:"explode"`
	AllowReserved bool    `json:"allowReserved"`
}
