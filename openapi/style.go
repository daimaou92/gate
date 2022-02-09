package openapi

import "fmt"

type Style string

const (
	STYLE_INVALID         Style = ""
	STYLE_MATRIX          Style = "matrix"
	STYLE_LABEL           Style = "label"
	STYLE_FORM            Style = "form"
	STYLE_SIMPLE          Style = "simple"
	STYLE_SPACE_DELIMITED Style = "spaceDelimited"
	STYLE_PIPE_DELIMITED  Style = "pipeDelimited"
	STYLE_DEEP_OBJECT     Style = "deepObject"
)

func StyleFromString(s string) Style {
	switch s {
	case string(STYLE_MATRIX):
		return STYLE_MATRIX
	case string(STYLE_LABEL):
		return STYLE_LABEL
	case string(STYLE_FORM):
		return STYLE_FORM
	case string(STYLE_SIMPLE):
		return STYLE_SIMPLE
	case string(STYLE_SPACE_DELIMITED):
		return STYLE_SPACE_DELIMITED
	case string(STYLE_PIPE_DELIMITED):
		return STYLE_PIPE_DELIMITED
	case string(STYLE_DEEP_OBJECT):
		return STYLE_DEEP_OBJECT
	}
	return STYLE_INVALID
}

func (s Style) Assert() error {
	if StyleFromString(string(s)) == STYLE_INVALID {
		return wrapErr(fmt.Errorf("invalid \"Style\" value: \"%s\"", s))
	}
	return nil
}
