package openapi

import "fmt"

type In string

const (
	IN_INVALID In = ""
	IN_QUERY   In = "query"
	IN_HEADER  In = "header"
	IN_PATH    In = "path"
	IN_COOKIE  In = "cookie"
)

func InFromString(s string) In {
	switch s {
	case string(IN_QUERY):
		return IN_QUERY
	case string(IN_HEADER):
		return IN_HEADER
	case string(IN_PATH):
		return IN_PATH
	case string(IN_COOKIE):
		return IN_COOKIE
	}
	return IN_INVALID
}

func (i In) Assert() error {
	if InFromString(string(i)) == IN_INVALID {
		return wrapErr(fmt.Errorf("invalid \"In\" value: \"%s\"", i))
	}
	return nil
}
