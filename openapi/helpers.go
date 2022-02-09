package openapi

import (
	"fmt"
	"runtime"
	"strings"
)

func wrapErr(err error, msgs ...string) error {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	src := frame.Function
	s := strings.Join(append([]string{src}, msgs...), " -> ")
	return fmt.Errorf("%s -> %s", s, err.Error())
}
