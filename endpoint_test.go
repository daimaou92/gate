package gate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	json "github.com/goccy/go-json"

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

func (p testPld) ContentType() ContentType {
	return ContentTypeJSON
}

var (
	tstReqBody   = &testPld{}
	tstQueryBody = &QueryPayload{}
)

func TestEndpointHandle(t *testing.T) {
	type tt struct {
		name               string
		route              string
		url                string
		method             string
		handler            Handler
		requestPayloadType Payload
		requestPayload     Payload
		queryPayloadType   Payload
		router             *httprouter.Router
		output             Payload
		outStatus          int
	}

	var port = 4444

	getfunc := func(method string, router *httprouter.Router) func(string, httprouter.Handle) {
		switch method {
		case http.MethodGet:
			return router.GET
		case http.MethodPut:
			return router.PUT
		case http.MethodPost:
			return router.POST
		case http.MethodPatch:
			return router.PATCH
		case http.MethodOptions:
			return router.OPTIONS
		case http.MethodDelete:
			return router.DELETE
		case http.MethodHead:
			return router.HEAD
		default:
			t.Fatalf("invalid method: %s", method)
		}
		return nil
	}
	tsts := []tt{
		{
			name:               "valid",
			route:              "/:namevalid",
			url:                fmt.Sprintf("http://localhost:%d/paul?key=value", port),
			method:             http.MethodPost,
			requestPayloadType: tstReqBody,
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			queryPayloadType: tstQueryBody,
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				return &testPld{
					Key:   "a",
					Value: "b",
				}, nil
			},
			router: httprouter.New(),
			output: &testPld{
				Key:   "a",
				Value: "b",
			},
			outStatus: StatusOK,
		}, {
			name:               "Request Body Unmarshal Error",
			route:              "/:namerbue",
			url:                fmt.Sprintf("http://localhost:%d/paul?key=value", port),
			method:             http.MethodPost,
			requestPayloadType: tstReqBody,
			requestPayload:     NewString("hi there"),
			queryPayloadType:   tstQueryBody,
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				return &testPld{
					Key:   "a",
					Value: "b",
				}, nil
			},
			router:    httprouter.New(),
			output:    nil,
			outStatus: StatusBadRequest,
		}, {
			name:               "Request Body Missing Error",
			route:              "/:namerbme",
			url:                fmt.Sprintf("http://localhost:%d/paul?key=value", port),
			method:             http.MethodPost,
			requestPayloadType: tstReqBody,
			queryPayloadType:   tstQueryBody,
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				return &testPld{
					Key:   "a",
					Value: "b",
				}, nil
			},
			router:    httprouter.New(),
			output:    nil,
			outStatus: StatusBadRequest,
		}, {
			name:               "Request Query Unmarshal Error",
			route:              "/:idrque",
			url:                fmt.Sprintf("http://localhost:%d/paul?key=value", port),
			method:             http.MethodPost,
			requestPayloadType: tstReqBody,
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			queryPayloadType: NewInt(0),
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				log.Println(rd.QueryParams)
				return &testPld{
					Key:   "a",
					Value: "b",
				}, nil
			},
			router:    httprouter.New(),
			output:    nil,
			outStatus: StatusBadRequest,
		}, {
			name:               "Request Query Empty Error",
			route:              "/:identifierrqee",
			url:                fmt.Sprintf("http://localhost:%d/paul", port),
			method:             http.MethodPost,
			requestPayloadType: tstReqBody,
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			queryPayloadType: tstQueryBody,
			handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
				log.Println(rd.QueryParams)
				return &testPld{
					Key:   "a",
					Value: "b",
				}, nil
			},
			router:    httprouter.New(),
			output:    nil,
			outStatus: StatusBadRequest,
		},
	}

	for _, tst := range tsts {
		t.Run(tst.name, func(t *testing.T) {
			ep := &endpoint{}
			var ps []Payload
			if tst.requestPayloadType != nil {
				ps = append(ps, tst.requestPayloadType)
			}

			if tst.queryPayloadType != nil {
				ps = append(ps, tst.queryPayloadType)
			}
			ep.update(tst.method, tst.route, tst.handler, ps...)
			ep.handle(getfunc(tst.method, tst.router))
			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: tst.router,
			}
			go func() {
				if err := server.ListenAndServe(); err != nil {
					if err != http.ErrServerClosed {
						log.Println(wrapErr(err))
					}
				}
			}()
			var (
				req *http.Request
				err error
			)
			if tst.requestPayload != nil {
				bs, _ := tst.requestPayload.Marshal()
				req, err = http.NewRequest(tst.method, tst.url, bytes.NewBuffer(bs))
			} else {
				req, err = http.NewRequest(tst.method, tst.url, nil)
			}
			if err != nil {
				server.Shutdown(context.TODO())
				t.Fatal(err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				server.Shutdown(context.TODO())
				t.Fatal(err)
			}
			if res.StatusCode != tst.outStatus {
				server.Shutdown(context.TODO())
				t.Fatalf("received code: %d. Wanted: %d", res.StatusCode, tst.outStatus)
			}

			if tst.output != nil {
				bs, err := io.ReadAll(res.Body)
				if err != nil {
					server.Shutdown(context.TODO())
					t.Fatal(err)
				}
				defer res.Body.Close()
				tbs, _ := tst.output.Marshal()
				if !bytes.Equal(bs, tbs) {
					server.Shutdown(context.TODO())
					t.Fatalf("wanted: %s\nGot: %s\n", tbs, bs)
				}
			}
			server.Shutdown(context.TODO())
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
