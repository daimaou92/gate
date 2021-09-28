package gate

import (
	"encoding/json"
	"fmt"
)

type String string

func NewString(s string) *String {
	v := String(s)
	return &v
}

func (s String) String() string {
	return string(s)
}

func (s String) Marshal() ([]byte, error) {
	bs, err := json.Marshal(s.String())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (s *String) Unmarshal(src []byte) error {
	var v string
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*s = String(v)
	return nil
}

func (String) ContentType() ContentType {
	return ContentTypeJSON
}

type Int8 int8

func NewInt8(i int8) *Int8 {
	v := Int8(i)
	return &v
}

func (i Int8) Int8() int8 {
	return int8(i)
}

func (i Int8) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Int8())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Int8) Unmarshal(src []byte) error {
	var v int8
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*i = Int8(v)
	return nil
}

func (Int8) ContentType() ContentType {
	return ContentTypeJSON
}

type Int int

func NewInt(i int) *Int {
	v := Int(i)
	return &v
}

func (i Int) Int() int {
	return int(i)
}

func (i Int) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Int())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Int) Unmarshal(src []byte) error {
	var v int
	if err := json.Unmarshal(src, &i); err != nil {
		return wrapErr(err)
	}
	*i = Int(v)
	return nil
}
func (Int) ContentType() ContentType {
	return ContentTypeJSON
}

type Int64 int64

func NewInt64(i int64) *Int64 {
	v := Int64(i)
	return &v
}

func (i Int64) Int64() int64 {
	return int64(i)
}

func (i Int64) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Int64())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Int64) Unmarshal(src []byte) error {
	var v int64
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*i = Int64(v)
	return nil
}

func (Int64) ContentType() ContentType {
	return ContentTypeJSON
}

type Uint8 uint8

func NewUint8(u uint8) *Uint8 {
	v := Uint8(u)
	return &v
}

func (u Uint8) Uint8() uint8 {
	return uint8(u)
}

func (i Uint8) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Uint8())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Uint8) Unmarshal(src []byte) error {
	var v uint8
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*i = Uint8(v)
	return nil
}

func (Uint8) ContentType() ContentType {
	return ContentTypeJSON
}

type Uint uint

func NewUint(u uint) *Uint {
	v := Uint(u)
	return &v
}

func (u Uint) Uint() uint {
	return uint(u)
}

func (i Uint) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Uint())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Uint) Unmarshal(src []byte) error {
	var v uint
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*i = Uint(v)
	return nil
}

func (Uint) ContentType() ContentType {
	return ContentTypeJSON
}

type Uint64 uint64

func NewUint64(u uint64) *Uint64 {
	v := Uint64(u)
	return &v
}

func (u Uint64) Uint64() uint64 {
	return uint64(u)
}

func (i Uint64) Marshal() ([]byte, error) {
	bs, err := json.Marshal(i.Uint64())
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (i *Uint64) Unmarshal(src []byte) error {
	var v uint64
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*i = Uint64(v)
	return nil
}

func (Uint64) ContentType() ContentType {
	return ContentTypeJSON
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

func (b Bool) Bool() bool {
	return bool(b)
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
	bs, err := json.Marshal(b)
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (b *Bool) Unmarshal(src []byte) error {
	var v bool
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*b = Bool(v)
	return nil
}

func (Bool) ContentType() ContentType {
	return ContentTypeJSON
}

type HTML string

func (h *HTML) Unmarshal(src []byte) error {
	var v string
	if err := json.Unmarshal(src, &v); err != nil {
		return wrapErr(err)
	}
	*h = HTML(v)
	return nil
}

func (h HTML) Marshal() ([]byte, error) {
	bs, err := json.Marshal(string(h))
	if err != nil {
		return nil, wrapErr(err)
	}
	return bs, nil
}

func (HTML) ContentType() ContentType {
	return "text/html; charset=utf-8"
}
