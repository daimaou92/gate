package openapi

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/mail"
	"reflect"
	"regexp"
)

type Type string

const (
	TYP_INVALID Type = ""
	TYP_INTEGER Type = "interger"
	TYP_NUMBER  Type = "number"
	TYP_STRING  Type = "string"
	TYP_BOOL    Type = "boolean"
	TYP_ARRAY   Type = "array"
	TYP_OBJECT  Type = "object"
	TYP_NULL    Type = "null"
)

func TypeFromString(s string) Type {
	switch s {
	case string(TYP_INTEGER):
		return TYP_INTEGER
	case string(TYP_NUMBER):
		return TYP_NUMBER
	case string(TYP_STRING):
		return TYP_STRING
	case string(TYP_BOOL):
		return TYP_BOOL
	case string(TYP_ARRAY):
		return TYP_ARRAY
	case string(TYP_OBJECT):
		return TYP_OBJECT
	case string(TYP_NULL):
		return TYP_NULL
	}
	return TYP_INVALID
}

func (t Type) Assert() error {
	if TypeFromString(string(t)) == TYP_INVALID {
		return wrapErr(fmt.Errorf("invalid type: \"%s\"", t))
	}
	return nil
}

func (t Type) AssertFormat(f Format) error {
	if m1, ok := typFmtMap[t]; ok {
		if _, ok := m1[f]; ok {
			return nil
		}
	} else {
		// Format is undefined for this type
		return nil
	}
	return wrapErr(fmt.Errorf("type format mismatch"))
}

func (t Type) AssertValue(v interface{}) error {
	// TODO: daimaou92
	return nil
}

type Format string

const (
	FMT_INVALID  Format = ""
	FMT_INT32    Format = "int32"
	FMT_INT64    Format = "int64"
	FMT_FLOAT    Format = "float"
	FMT_DOUBLE   Format = "double"
	FMT_BYTE     Format = "byte"
	FMT_BINARY   Format = "binary"
	FMT_DATE     Format = "date"
	FMT_DATETIME Format = "date-time"
	FMT_PWD      Format = "password"
	FMT_EMAIL    Format = "email"
	FMT_HOSTNAME Format = "hostname"
	FMT_IPV4     Format = "ipv4"
	FMT_IPV6     Format = "ipv6"
	FMT_URI      Format = "uri"
	FMT_URIREF   Format = "uriref"
)

func FormatFromString(s string) Format {
	switch s {
	case string(FMT_INT32):
		return FMT_INT32
	case string(FMT_INT64):
		return FMT_INT64
	case string(FMT_FLOAT):
		return FMT_FLOAT
	case string(FMT_DOUBLE):
		return FMT_DOUBLE
	case string(FMT_BYTE):
		return FMT_BYTE
	case string(FMT_BINARY):
		return FMT_BINARY
	case string(FMT_DATE):
		return FMT_DATE
	case string(FMT_DATETIME):
		return FMT_DATETIME
	case string(FMT_PWD):
		return FMT_PWD
	case string(FMT_EMAIL):
		return FMT_EMAIL
	case string(FMT_HOSTNAME):
		return FMT_HOSTNAME
	case string(FMT_IPV4):
		return FMT_IPV4
	case string(FMT_IPV6):
		return FMT_IPV6
	case string(FMT_URI):
		return FMT_URI
	case string(FMT_URIREF):
		return FMT_URIREF
	}
	return FMT_INVALID
}

func (f Format) Assert() error {
	if FormatFromString(string(f)) == FMT_INVALID {
		return wrapErr(fmt.Errorf("invalid format: \"%s\"", f))
	}
	return nil
}

func makeInt(v interface{}) (int, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int:
		i := v.(int)
		return i, nil
	case reflect.Int8:
		i := v.(int8)
		return int(i), nil
	case reflect.Int16:
		i := v.(int16)
		return int(i), nil
	case reflect.Int32:
		i := v.(int32)
		return int(i), nil
	case reflect.Int64:
		i := v.(int64)
		return int(i), nil
	case reflect.Uint:
		i := v.(uint)
		if i <= uint(math.MaxInt) {
			return int(i), nil
		} else {
			return 0, wrapErr(fmt.Errorf("unsigned int value out of bound for int"))
		}
	case reflect.Uint8:
		i := v.(uint8)
		return int(i), nil
	case reflect.Uint16:
		i := v.(uint16)
		return int(i), nil
	case reflect.Uint32:
		i := v.(uint32)
		return int(i), nil
	case reflect.Uint64:
		i := v.(uint64)
		if i <= uint64(math.MaxInt) {
			return int(i), nil
		} else {
			return 0, wrapErr(fmt.Errorf("unsigned int64 value out of bound for int"))
		}
	}
	return 0, wrapErr(fmt.Errorf("no int found"))
}

func assertInt32(v interface{}) error {
	if _, ok := v.(int32); ok {
		return nil
	}
	i, err := makeInt(v)
	if err != nil {
		return wrapErr(err)
	}
	if i <= math.MaxInt32 && i >= math.MinInt32 {
		return nil
	}
	return wrapErr(fmt.Errorf("value out of bounds for int32"))
}

func assertFloat(v interface{}) error {
	if _, ok := v.(float32); !ok {
		if _, ok := v.(float64); !ok {
			return wrapErr(fmt.Errorf("not a float"))
		}
	}
	return nil
}

func assertByte(v interface{}) error {
	var s string
	if b, ok := v.([]byte); !ok {
		if t, ok := v.(string); !ok {
			return wrapErr(fmt.Errorf("not []byte or string"))
		} else {
			s = t
		}
	} else {
		s = string(b)
	}
	if _, err := base64.StdEncoding.DecodeString(s); err != nil {
		return wrapErr(err, "not base64 encoded")
	}
	return nil
}

func assertBinary(v interface{}) error {
	if _, ok := v.([]uint8); !ok {
		if _, ok := v.([]byte); !ok {
			if _, ok := v.(string); !ok {
				if _, ok := v.([]rune); !ok {
					return wrapErr(fmt.Errorf("not a binary"))
				}
			}
		}
	}
	return nil
}

func assertInt64(v interface{}) error {
	if _, ok := v.(int64); ok {
		return nil
	}
	if _, err := makeInt(v); err != nil {
		return wrapErr(err)
	}
	// All ints are valid int64 as well
	return nil
}

func getString(v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		t, ok := v.([]byte)
		if !ok {
			return "", wrapErr(fmt.Errorf("not a string"))
		}
		s = string(t)
	}
	return s, nil
}

func rfc3339Regex(s string) (string, error) {
	timeHr := `[0-1][0-9]|2[0-3]`
	timeMin := `[0-5][0-9]`
	timeSec := `[0-5][0-9]|60`
	timeSecFrac := `\.[0-9]+`
	partialTime := fmt.Sprintf(
		`(%s):(%s):(%s)(%s)?`,
		timeHr, timeMin, timeSec, timeSecFrac,
	)
	timeNumOffset := fmt.Sprintf(
		`(+|-)%s:%s`,
		timeHr, timeMin,
	)
	timeOffset := fmt.Sprintf(
		`(z|Z)|%s`, timeNumOffset,
	)
	fullTime := fmt.Sprintf(
		`%s%s`,
		partialTime, timeOffset,
	)
	fullDate := `[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])`
	dateTime := fmt.Sprintf(
		`%s(t|T)%s`,
		fullDate, fullTime,
	)
	switch s {
	case "full-date":
		return fullDate, nil
	case "date-time":
		return dateTime, nil
	}
	return "", fmt.Errorf("unhandled \"%s\"", s)
}

// RFC 3339 full-date
func assertDate(v interface{}) error {
	s, err := getString(v)
	if err != nil {
		return wrapErr(err)
	}
	rs, err := rfc3339Regex("full-date")
	if err != nil {
		return wrapErr(err)
	}
	r, err := regexp.Compile("^" + rs + "$")
	if err != nil {
		return wrapErr(err, "regexp.Compile failed")
	}
	if !r.Match([]byte(s)) {
		return wrapErr(fmt.Errorf(
			"invalid date format, refer, full-date, $%s",
			"https://xml2rfc.tools.ietf.org/public/rfc/html/rfc3339.html#anchor14",
		))
	}
	return nil
}

// RFC 3339 date-time
func assertDateTime(v interface{}) error {
	s, err := getString(v)
	if err != nil {
		return wrapErr(err)
	}
	rs, err := rfc3339Regex("date-time")
	if err != nil {
		return wrapErr(err)
	}
	r, err := regexp.Compile("$" + rs + "$")
	if err != nil {
		return wrapErr(err, "regexp.MustCompile failed")
	}
	if !r.Match([]byte(s)) {
		return wrapErr(fmt.Errorf(
			"invalid date format, refer, date-time, $%s",
			"https://xml2rfc.tools.ietf.org/public/rfc/html/rfc3339.html#anchor14",
		))
	}
	return nil
}

func assertEmail(v interface{}) error {
	s, err := getString(v)
	if err != nil {
		return wrapErr(err)
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return wrapErr(err, "not a valid email")
	}
	return nil
}

func (f Format) AssertValue(v interface{}) error {
	switch f {
	case FMT_INT32:
		if err := assertInt32(v); err != nil {
			return wrapErr(err)
		}
	case FMT_INT64:
		if err := assertInt64(v); err != nil {
			return wrapErr(err)
		}
	case FMT_FLOAT, FMT_DOUBLE:
		if err := assertFloat(v); err != nil {
			return wrapErr(err)
		}
	case FMT_BYTE:
		if err := assertByte(v); err != nil {
			return wrapErr(err)
		}
	case FMT_BINARY:
		if err := assertBinary(v); err != nil {
			return wrapErr(err)
		}
	case FMT_DATE:
		if err := assertDate(v); err != nil {
			return wrapErr(err)
		}
	case FMT_DATETIME:
		if err := assertDateTime(v); err != nil {
			return wrapErr(err)
		}
	case FMT_PWD:
		if _, err := getString(v); err != nil {
			return wrapErr(err)
		}
	case FMT_EMAIL:
		if err := assertEmail(v); err != nil {
			return wrapErr(err)
		}

		// TODO: daimaou92

	}
	return nil
}

var typFmtMap = map[Type]map[Format]bool{
	TYP_INTEGER: {
		FMT_INT32: true,
		FMT_INT64: true,
	},
	TYP_NUMBER: {
		FMT_INT32:  true,
		FMT_INT64:  true,
		FMT_FLOAT:  true,
		FMT_DOUBLE: true,
	},
	TYP_STRING: {
		FMT_BYTE:     true,
		FMT_BINARY:   true,
		FMT_DATE:     true,
		FMT_DATETIME: true,
		FMT_PWD:      true,
		FMT_EMAIL:    true,
		FMT_HOSTNAME: true,
		FMT_IPV4:     true,
		FMT_IPV6:     true,
		FMT_URI:      true,
		FMT_URIREF:   true,
	},
}
