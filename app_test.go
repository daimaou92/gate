package gate

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestListen(t *testing.T) {
	type tt struct {
		name       string
		ao         AppOptions
		method     string
		url        string
		route      string
		reqBody    Payload
		handler    Handler
		resBody    Payload
		statusCode int
	}

	tsts := []tt{
		{
			name:   http.MethodGet,
			method: http.MethodGet,
			url:    "http://localhost:5050/sarkar",
			route:  "/:name",
			ao: AppOptions{
				Addr: ":5050",
				Info: openapi3.Info{
					Title:   "test api",
					Version: "0.0.0",
				},
			},
			resBody: &testPld{
				Key:   "name",
				Value: "sarkar",
			},
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				n := rd.Params.ByName("name")
				return &testPld{
					Key:   "name",
					Value: n,
				}, nil
			},
			statusCode: StatusOK,
		}, {
			name:   http.MethodPost,
			method: http.MethodPost,
			url:    "http://localhost:4747/sarkar",
			route:  "/:name",
			ao: AppOptions{
				Addr: ":4747",
				Info: openapi3.Info{
					Title:   "test api",
					Version: "0.0.0",
				},
			},
			reqBody: NewString("This is post body"),
			resBody: NewString("This is post body - sarkar"),
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				j, ok := rd.Body.(*String)
				if !ok {
					return nil, ErrBadRequest
				}
				n := rd.Params.ByName("name")
				r := fmt.Sprintf("%s - %s", j.String(), n)
				return NewString(r), nil
			},
			statusCode: StatusOK,
		}, {
			name:   http.MethodPatch,
			method: http.MethodPatch,
			url:    "http://localhost:2987/sarkar",
			route:  "/:name",
			ao: AppOptions{
				Addr: ":2987",
				Info: openapi3.Info{
					Title:   "test api",
					Version: "0.0.0",
				},
			},
			reqBody: NewString("This is post body"),
			resBody: NewString("This is post body - sarkar"),
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				j, ok := rd.Body.(*String)
				if !ok {
					return nil, ErrBadRequest
				}
				n := rd.Params.ByName("name")
				r := fmt.Sprintf("%s - %s", j.String(), n)
				return NewString(r), nil
			},
			statusCode: StatusOK,
		}, {
			name:   http.MethodPut,
			method: http.MethodPut,
			url:    "http://localhost:8678/sarkar",
			route:  "/:name",
			ao: AppOptions{
				Addr: ":8678",
				Info: openapi3.Info{
					Title:   "test api",
					Version: "0.0.0",
				},
			},
			reqBody: NewString("This is post body"),
			resBody: NewString("This is post body - sarkar"),
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				j, ok := rd.Body.(*String)
				if !ok {
					return nil, ErrBadRequest
				}
				n := rd.Params.ByName("name")
				r := fmt.Sprintf("%s - %s", j.String(), n)
				return NewString(r), nil
			},
			statusCode: StatusOK,
		}, {
			name:   http.MethodDelete,
			method: http.MethodDelete,
			url:    "http://localhost:9999/sarkar",
			route:  "/:name",
			ao: AppOptions{
				Addr: ":9999",
				Info: openapi3.Info{
					Title:   "test api",
					Version: "0.0.0",
				},
			},
			reqBody: NewString("This is post body"),
			resBody: NewString("This is post body - sarkar"),
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				j, ok := rd.Body.(*String)
				if !ok {
					return nil, ErrBadRequest
				}
				n := rd.Params.ByName("name")
				r := fmt.Sprintf("%s - %s", j.String(), n)
				return NewString(r), nil
			},
			statusCode: StatusOK,
		},
	}

	for _, tst := range tsts {
		t.Run(tst.name, func(t *testing.T) {
			app, err := New(tst.ao)
			if err != nil {
				t.Fatal(err)
			}
			var f HandleFuncType
			switch tst.method {
			case http.MethodGet:
				f = app.Get
			case http.MethodPost:
				f = app.Post
			case http.MethodDelete:
				f = app.Delete
			case http.MethodHead:
				f = app.Head
			case http.MethodOptions:
				f = app.Options
			case http.MethodPut:
				f = app.Put
			case http.MethodPatch:
				f = app.Patch
			default:
				log.Fatalf("invalid method: %s", tst.method)
			}
			var args []Payload
			if tst.reqBody != nil {
				args = append(args, tst.reqBody)
			}
			f(tst.route, tst.handler, args...)

			// Start server
			go func() {
				if err := app.Listen(); err != nil {
					log.Println(wrapErr(err, "listen failed"))
				}
			}()

			var (
				r *http.Request
			)
			if tst.reqBody != nil {
				bs, err := tst.reqBody.Marshal()
				if err != nil {
					t.Fatalf("marshal failed: %s", err.Error())
				}
				r, err = http.NewRequest(tst.method, tst.url, bytes.NewBuffer(bs))
				if err != nil {
					t.Fatalf("newrequest failed: %s", err.Error())
				}
			} else {
				var err error
				r, err = http.NewRequest(tst.method, tst.url, nil)
				if err != nil {
					t.Fatalf("newrequest, emptybody failed: %s", err.Error())
				}
			}
			res, err := http.DefaultClient.Do(r)
			if err != nil {
				log.Printf("making client request faild: %s", err.Error())
			}
			defer res.Body.Close()

			if res.StatusCode != tst.statusCode {
				t.Fatalf("statuscode wanted: %d. got %d", tst.statusCode, res.StatusCode)
			}
			if tst.resBody != nil {
				bs, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("res body readall failed: %s", err.Error())
				}
				bsw, err := tst.resBody.Marshal()
				if err != nil {
					t.Fatalf("tst.resBody marshal failed: %s", err.Error())
				}

				if !bytes.Equal(bsw, bs) {
					t.Fatalf("wanted %s. got %s", bsw, bs)
				}
			}
		})
	}
}
