package gate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
)

type testPld struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p testPld) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *testPld) Unmarshal(src []byte) error {
	var v testPld
	if err := json.Unmarshal(src, &v); err != nil {
		return err
	}
	*p = v
	return nil
}
func TestEndpointHandle(t *testing.T) {
	type tsts struct {
		name    string
		method  string
		route   string
		reqBody Payload
		resBody Payload
		resCode int
		handler Handler
		params  httprouter.Params
	}

	tt := []tsts{
		{
			name:    "Get",
			method:  http.MethodGet,
			route:   "/:name",
			resBody: NewString("Hi John!"),
			resCode: StatusOK,
			params: httprouter.Params{httprouter.Param{
				Key:   "name",
				Value: "John",
			}},
			handler: func(rc *RequestCtx, p Payload) (Payload, error) {
				n := rc.Params.ByName("name")
				s := new(String)
				*s = String(fmt.Sprintf("Hi %s!", n))
				return s, nil
			},
		}, {
			name:   "Post",
			method: http.MethodPost,
			route:  "/:name",
			reqBody: &testPld{
				Key:   "key",
				Value: "value",
			},
			resBody: &testPld{
				Key:   "key-executed",
				Value: "value-executed",
			},
			resCode: StatusOK,
			params: httprouter.Params{httprouter.Param{
				Key:   "name",
				Value: "John",
			}},
			handler: func(rc *RequestCtx, p Payload) (Payload, error) {
				bs, err := p.Marshal()
				if err != nil {
					return nil, ErrBadRequest
				}
				tp := testPld{
					Key:   "key",
					Value: "value",
				}
				bst, err := tp.Marshal()
				if err != nil {
					return nil, ErrInternalServerError
				}
				if !bytes.Equal(bst, bs) {
					log.Printf("wanted: %s\nGot: %s\n", bst, bs)
					return nil, ErrInternalServerError
				}
				return &testPld{
					Key:   "key-executed",
					Value: "value-executed",
				}, nil
			},
		}, {
			name:   "Post",
			method: http.MethodPost,
			route:  "/:name",
			reqBody: &testPld{
				Key:   "key",
				Value: "value",
			},
			resBody: &testPld{
				Key:   "key-executed",
				Value: "value-executed",
			},
			resCode: StatusOK,
			params: httprouter.Params{httprouter.Param{
				Key:   "name",
				Value: "John",
			}},
			handler: func(rc *RequestCtx, p Payload) (Payload, error) {
				bs, err := p.Marshal()
				if err != nil {
					return nil, ErrBadRequest
				}
				tp := testPld{
					Key:   "key",
					Value: "value",
				}
				bst, err := tp.Marshal()
				if err != nil {
					return nil, ErrInternalServerError
				}
				if !bytes.Equal(bst, bs) {
					log.Printf("wanted: %s\nGot: %s\n", bst, bs)
					return nil, ErrInternalServerError
				}
				return &testPld{
					Key:   "key-executed",
					Value: "value-executed",
				}, nil
			},
		},
	}

	for _, tst := range tt {
		t.Run(tst.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tst.method, tst.route, nil)
			if tst.reqBody != nil {
				bs, err := tst.reqBody.Marshal()
				if err != nil {
					t.Fatalf("Paylaod.marshal failed: %s\n", err.Error())
				}
				req = httptest.NewRequest(tst.method, tst.route, bytes.NewBuffer(bs))
			}

			ep := new(endpoint)
			ep.update(tst.route, tst.reqBody, tst.handler, rr, req, tst.params)
			ep.handle()
			if tst.resCode != rr.Code {
				t.Fatalf("expected code: %d. got: %d\n", tst.resCode, rr.Code)
			}
			bs, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Fatalf(err.Error())
			}
			defer rr.Result().Body.Close()
			bst, err := tst.resBody.Marshal()
			if err != nil {
				t.Fatalf(err.Error())
			}
			if !bytes.Equal(bst, bs) {
				t.Fatalf("expected: %s\nGot: %s\n", bst, bs)
			}
		})
	}
}

type testRW string

func (trw testRW) Header() http.Header {
	return http.Header{}
}

func (trw testRW) Write(bs []byte) (int, error) {
	log.Println(bs)
	return 0, nil
}

func (trw testRW) WriteHeader(statusCode int) {}

func TestEndpointReset(t *testing.T) {
	type tsts struct {
		name string
		ep   *endpoint
	}

	s := ""
	tt := []tsts{
		{
			name: "default",
			ep: &endpoint{
				route: "/",
				handler: func(rc *RequestCtx, p Payload) (Payload, error) {
					return nil, nil
				},
				Payload: &testPld{},
				rw:      testRW(""),
				r:       &http.Request{},
				params:  httprouter.Params{},
				typ:     reflect.TypeOf(s),
				val:     reflect.ValueOf(s),
			},
		},
	}

	for _, tst := range tt {
		t.Run(tst.name, func(t *testing.T) {
			tst.ep.reset()
			if tst.ep.route != "" {
				t.Fatalf("route not empty")
			}
			if tst.ep.handler != nil {
				t.Fatalf("handler not nil")
			}

			if tst.ep.Payload != nil {
				t.Fatalf("payload not nil")
			}
			if tst.ep.rw != nil {
				t.Fatalf("rw not nil")
			}

			if tst.ep.r != nil {
				t.Fatalf("request not nil")
			}

			if len(tst.ep.params) != 0 {
				t.Fatalf("params not nil")
			}
			if tst.ep.typ != nil {
				t.Fatalf("typ not nil")
			}
		})
	}
}
