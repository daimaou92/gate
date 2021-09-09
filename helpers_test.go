package gate

import (
	"fmt"
	"testing"
)

func TestWrapErr(t *testing.T) {
	p := "test error"
	o := fmt.Sprintf("github.com/enalk-com/gate.TestWrapErr -> %s", p)
	v := wrapErr(fmt.Errorf(p)).Error()
	if o != v {
		t.Fatalf("wanted: %s. got: %s", o, v)
	}
}
