package gate

import (
	"strings"
)

type GateErr error

type Error struct {
	Code    int
	Message []string
}

func (e *Error) Error() string {
	return strings.Join(e.Message, "\n")
}

func NewError(code int, message ...string) *Error {
	e := &Error{
		Code:    code,
		Message: message,
	}
	if len(message) == 0 {
		if m, ok := httpStatusMessage[e.Code]; ok {
			e.Message = append(e.Message, m)
		}
	}
	return e
}
