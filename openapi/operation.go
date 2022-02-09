package openapi

type Operation struct {
	OperationID  string        `json:"operationId"`
	Tags         []string      `json:"tags"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	Parameters   interface{}   `json:"parameters"` // ParameterObject | ReferenceObject
}
