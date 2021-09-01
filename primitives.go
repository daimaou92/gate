package gate

import (
	"fmt"
	"strconv"
)

type String string

func NewString(s string) *String {
	v := String(s)
	return &v
}

func (s String) Marshal() ([]byte, error) {
	return []byte(s), nil
}

func (s *String) Unmarshal(src []byte) error {
	*s = String(string(src))
	return nil
}

type Int8 int8

func NewInt8(i int8) *Int8 {
	v := Int8(i)
	return &v
}

func (i Int8) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Int8) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Int8(v)
	return nil
}

type Int int

func NewInt(i int) *Int {
	v := Int(i)
	return &v
}

func (i Int) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Int) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Int(v)
	return nil
}

type Int64 int64

func NewInt64(i int64) *Int64 {
	v := Int64(i)
	return &v
}

func (i Int64) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Int64) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Int64(v)
	return nil
}

type Uint8 uint8

func NewUint8(u uint8) *Uint8 {
	v := Uint8(u)
	return &v
}

func (i Uint8) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Uint8) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Uint8(v)
	return nil
}

type Uint uint

func NewUint(u uint) *Uint {
	v := Uint(u)
	return &v
}

func (i Uint) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Uint) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Uint(v)
	return nil
}

type Uint64 uint64

func NewUint64(u uint64) *Uint64 {
	v := Uint64(u)
	return &v
}

func (i Uint64) Marshal() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Uint64) Unmarshal(src []byte) error {
	v, err := strconv.Atoi(string(src))
	if err != nil {
		return err
	}
	*i = Uint64(v)
	return nil
}

type Bool bool

func NewBool(b bool) *Bool {
	v := Bool(b)
	return &v
}

func (b Bool) string() string {
	if b {
		return "true"
	}
	return "false"
}

func (b *Bool) fromString(s string) error {
	if b == nil {
		return fmt.Errorf("[ERR] no memory allocated")
	}
	switch s {
	case "true":
		*b = true
		return nil
	case "false":
		*b = false
		return nil
	}
	return fmt.Errorf("[ERR] invalid input: %s", s)
}

func (b Bool) Marshal() ([]byte, error) {
	return []byte(b.string()), nil
}

func (b *Bool) Unmarshal(src []byte) error {
	return b.fromString(string(src))
}
