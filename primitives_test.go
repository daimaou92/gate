package gate

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPMarshal(t *testing.T) {
	type tt struct {
		name   string
		input  Payload
		output []byte
	}

	tsts := []tt{
		{
			name:   "string",
			input:  NewString("hi there"),
			output: []byte("hi there"),
		}, {
			name:   "int8",
			input:  NewInt8(98),
			output: []byte("98"),
		}, {
			name:   "int",
			input:  NewInt(100),
			output: []byte("100"),
		}, {
			name:   "int64",
			input:  NewInt64(678987986876),
			output: []byte("678987986876"),
		}, {
			name:   "uint8",
			input:  NewUint8(56),
			output: []byte("56"),
		}, {
			name:   "uint",
			input:  NewUint(54),
			output: []byte("54"),
		}, {
			name:   "uint64",
			input:  NewUint64(45),
			output: []byte("45"),
		}, {
			name:   "bool",
			input:  NewBool(true),
			output: []byte("true"),
		},
	}

	for _, tst := range tsts {
		t.Run(tst.name, func(t *testing.T) {
			bs, err := tst.input.Marshal()
			if err != nil {
				t.Fatalf(err.Error())
			}
			if !bytes.Equal(bs, tst.output) {
				t.Fatalf("wanted: %s. got %s", tst.output, bs)
			}
		})
	}
}

func TestPUnmarshal(t *testing.T) {
	type tt struct {
		name   string
		obj    Payload
		input  []byte
		output error
	}

	tsts := []tt{
		{
			name:   "string",
			obj:    NewString(""),
			input:  []byte("hi there"),
			output: nil,
		}, {
			name:   "int8",
			obj:    NewInt8(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "int8err",
			obj:    NewInt8(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "int",
			obj:    NewInt(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "interr",
			obj:    NewInt(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "int64",
			obj:    NewInt64(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "int64err",
			obj:    NewInt64(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "uint8",
			obj:    NewUint8(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "uint8err",
			obj:    NewUint8(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "uint",
			obj:    NewUint(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "uinterr",
			obj:    NewUint(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "uint64",
			obj:    NewUint64(0),
			input:  []byte("90"),
			output: nil,
		}, {
			name:   "uint64err",
			obj:    NewUint64(0),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		}, {
			name:   "bool",
			obj:    NewBool(false),
			input:  []byte("true"),
			output: nil,
		}, {
			name:   "boolerr",
			obj:    NewBool(false),
			input:  []byte("hello world"),
			output: fmt.Errorf(""),
		},
	}

	for _, tst := range tsts {
		t.Run(tst.name, func(t *testing.T) {
			if err := tst.obj.Unmarshal(tst.input); err != nil {
				if tst.output != nil {
					return
				}
				t.Fatalf(err.Error())
			}
		})
	}
}
