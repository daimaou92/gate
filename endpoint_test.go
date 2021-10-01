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
		name           string
		ec             EndpointConfig
		url            string
		requestPayload Payload
		output         Payload
		outStatus      int
		routerFunc     func(string, httprouter.Handle)
	}

	var port = 4444
	router := httprouter.New()
	tsts := []tt{
		{
			name: "valid",
			ec: EndpointConfig{
				Path:    "/1/:namevalid",
				method:  http.MethodPost,
				Payload: NewEndpointPayload(tstReqBody, tstQueryBody),
				Handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
					return &testPld{
						Key:   "a",
						Value: "b",
					}, nil
				},
			},
			routerFunc: router.POST,
			url:        fmt.Sprintf("http://localhost:%d/1/paul?key=value", port),
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			output: &testPld{
				Key:   "a",
				Value: "b",
			},
			outStatus: StatusOK,
		}, {
			name: "Request Body Unmarshal Error",
			ec: EndpointConfig{
				Path:    "/2/:namerbue",
				Payload: NewEndpointPayload(tstReqBody, tstQueryBody),
				Handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
					return &testPld{
						Key:   "a",
						Value: "b",
					}, nil
				},
				method: http.MethodPost,
			},
			routerFunc:     router.POST,
			url:            fmt.Sprintf("http://localhost:%d/2/paul?key=value", port),
			requestPayload: NewString("hi there"),
			output:         nil,
			outStatus:      StatusBadRequest,
		}, {
			name: "Request Body Missing Error",
			ec: EndpointConfig{
				Path:    "/3/:namerbme",
				Payload: NewEndpointPayload(tstReqBody, tstQueryBody),
				Handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
					return &testPld{
						Key:   "a",
						Value: "b",
					}, nil
				},
				method: http.MethodPost,
			},
			routerFunc: router.POST,
			url:        fmt.Sprintf("http://localhost:%d/3/paul?key=value", port),
			output:     nil,
			outStatus:  StatusBadRequest,
		}, {
			name: "Request Query Unmarshal Error",
			ec: EndpointConfig{
				Path:    "/4/:idrque",
				Payload: NewEndpointPayload(tstReqBody, NewInt(0)),
				Handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
					log.Println(rd.QueryParams)
					return &testPld{
						Key:   "a",
						Value: "b",
					}, nil
				},
				method: http.MethodPost,
			},
			routerFunc: router.POST,
			url:        fmt.Sprintf("http://localhost:%d/4/paul?key=value", port),
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			output:    nil,
			outStatus: StatusBadRequest,
		}, {
			name: "Request Query Empty Error",
			ec: EndpointConfig{
				Path:    "/5/:identifierrqee",
				Payload: NewEndpointPayload(tstReqBody, tstQueryBody),
				Handler: func(rc *RequestCtx, rd *RequestData) (Payload, error) {
					log.Println(rd.QueryParams)
					return &testPld{
						Key:   "a",
						Value: "b",
					}, nil
				},
				method: http.MethodPost,
			},
			routerFunc: router.POST,
			url:        fmt.Sprintf("http://localhost:%d/5/paul", port),
			requestPayload: &testPld{
				Key:   "a",
				Value: "b",
			},
			output:    nil,
			outStatus: StatusBadRequest,
		},
	}

	for _, tst := range tsts {
		t.Run(tst.name, func(t *testing.T) {
			ep := tst.ec.endpoint()
			ep.handle(tst.routerFunc)

			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: router,
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
				req, err = http.NewRequest(tst.ec.method, tst.url, bytes.NewBuffer(bs))
			} else {
				req, err = http.NewRequest(tst.ec.method, tst.url, nil)
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
