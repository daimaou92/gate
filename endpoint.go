package gate

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"

	json "github.com/goccy/go-json"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/julienschmidt/httprouter"
)

var requestDataPool = sync.Pool{
	New: func() interface{} {
		return new(RequestData)
	},
}

type endpoint struct {
	method          string
	route           string
	handler         Handler
	requestPayload  Payload
	queryPayload    Payload
	responsePayload Payload
	mexclusions     []string
	requestPool     sync.Pool
	queryPool       sync.Pool
}

func (ep *endpoint) initPools() {
	if ep.requestPayload != nil {
		ep.requestPool = sync.Pool{
			New: func() interface{} {
				inpt := reflect.TypeOf(ep.requestPayload)
				switch inpt.Kind() {
				case reflect.Array, reflect.Chan,
					reflect.Map, reflect.Ptr, reflect.Slice:
					inpt = inpt.Elem()
				}
				val := reflect.ValueOf(ep.requestPayload)
				v := reflect.New(inpt).Elem()
				v.Set(val.Elem())
				return v.Addr()
			},
		}
	}

	if ep.queryPayload != nil {
		ep.queryPool = sync.Pool{
			New: func() interface{} {
				inpt := reflect.TypeOf(ep.queryPayload)
				switch inpt.Kind() {
				case reflect.Array, reflect.Chan,
					reflect.Map, reflect.Ptr, reflect.Slice:
					inpt = inpt.Elem()
				}
				val := reflect.ValueOf(ep.queryPayload)
				v := reflect.New(inpt).Elem()
				v.Set(val.Elem())
				return v.Addr()
			},
		}
	}
}

func (ep endpoint) routeDetails() (string, []string) {
	// qps := queryParams(ep.Payload)
	params := pathParams(ep.route)
	if len(params) == 0 {
		return ep.route, nil
	}
	r := ep.route
	for i, param := range params {
		fr := fmt.Sprintf("{%s}", strings.Trim(strings.Replace(param, ":", "", 1), "/"))
		r = strings.Replace(r, param, fr, 1)
		pname := strings.Trim(strings.Replace(param, ":", "", 1), "/")
		params[i] = pname
	}
	return r, params
}

func (ep endpoint) requestSchema() (openapi3.Schema, error) {
	s, err := schemaFromType(reflect.TypeOf(ep.handler).In(1))
	if err != nil {
		return *openapi3.NewSchema(), wrapErr(err)
	}
	return s, nil
}

func (ep endpoint) responseSchema() (openapi3.Schema, error) {
	t := reflect.TypeOf(ep.handler)
	if t.Kind() != reflect.Func {
		return openapi3.Schema{}, wrapErr(fmt.Errorf("type if not a func"))
	}
	s, err := schemaFromType(t.Out(0))
	if err != nil {
		return openapi3.Schema{}, wrapErr(err)
	}
	return s, nil
}

// func (ep endpoint) generatePathItem() {
// 	op := openapi3.NewOperation()
// 	op.OperationID = ep.route
// 	formattedRoute, params, queryParams := ep.routeDetails()

// }

// func (ep *endpoint) update(
// 	ec EndpointConfig,
// ) {
// 	ep.method = ec.method
// 	ep.route = ec.Route
// 	ep.handler = ec.Handler

// 	if len(ps) > 0 {
// 		pl := ps[0]
// 		if _, ok := requestBodyPool[route]; !ok {
// 			requestBodyPool[route] = sync.Pool{
// 				New: func() interface{} {
// 					inpt := reflect.TypeOf(pl)
// 					switch inpt.Kind() {
// 					case reflect.Array, reflect.Chan,
// 						reflect.Map, reflect.Ptr, reflect.Slice:
// 						inpt = inpt.Elem()
// 					}
// 					val := reflect.ValueOf(pl)
// 					v := reflect.New(inpt).Elem()
// 					v.Set(val.Elem())
// 					return v.Addr()
// 				},
// 			}
// 		}
// 		ep.requestPayload = pl
// 	}

// 	if len(ps) > 1 {
// 		pl := ps[1]
// 		if _, ok := queryPool[route]; !ok {
// 			queryPool[route] = sync.Pool{
// 				New: func() interface{} {
// 					inpt := reflect.TypeOf(pl)
// 					switch inpt.Kind() {
// 					case reflect.Array, reflect.Chan,
// 						reflect.Map, reflect.Ptr, reflect.Slice:
// 						inpt = inpt.Elem()
// 					}
// 					val := reflect.ValueOf(pl)
// 					v := reflect.New(inpt).Elem()
// 					v.Set(val.Elem())
// 					return v.Addr()
// 				},
// 			}
// 		}

// 		ep.queryPayload = pl
// 	}

// 	if len(ps) > 2 {
// 		ep.responsePayload = ps[2]
// 	}
// }

// func (ep *endpoint) reset() {
// 	ep.handler = nil
// 	// ep.Payload = nil
// 	ep.route = ""
// 	ep.method = ""
// 	ep.requestPayload = nil
// 	ep.queryPayload = nil
// 	ep.responsePayload = nil
// 	// ep.typ = nil
// 	// ep.val = reflect.Value{}
// }

func (ep *endpoint) handle(f func(string, httprouter.Handle)) {
	f(ep.route, func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		rd, ok := requestDataPool.Get().(*RequestData)
		if !ok {
			panic(wrapErr(fmt.Errorf("requestDataPool returned not *RequestData.... aaaaaaa")))
		}
		defer func() {
			rd.Custom = nil
			requestDataPool.Put(rd)
		}()
		rd.Custom = map[string]interface{}{}
		rd.Params = params

		badrequest := func(msg string) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(StatusBadRequest)
			w.Write([]byte(msg))
		}

		// Request Payload
		if ep.requestPayload != nil {
			pv := ep.requestPool.Get()
			if pv == nil {
				panic(wrapErr(fmt.Errorf("requestPayload Pool returned nil....aaaaaaaa")))
			}
			defer ep.requestPool.Put(pv)

			v, ok := pv.(reflect.Value)
			if !ok {
				panic(wrapErr(fmt.Errorf("requestBodyPool returned value of type not equal to reflect.Value")))
			}
			reflect.ValueOf(rd).Elem().FieldByName("Body").Set(v)

			bs, err := io.ReadAll(r.Body)
			if err != nil {
				// log.Println(wrapErr(err, "readall failed"))
				if err != io.EOF {
					log.Println(wrapErr(err))
					badrequest("connection error")
					return
				}
			}

			if len(bs) > 0 {
				if err := rd.Body.Unmarshal(bs); err != nil {
					log.Println(wrapErr(err, "request unmarshal failed"))
					badrequest("invalid payload")
					return
				}
			} else {
				badrequest("empty payload")
				return
			}
		}

		// Query Params
		if ep.queryPayload != nil {
			pv := ep.queryPool.Get()
			if pv == nil {
				panic(wrapErr(fmt.Errorf("query pool returned empty....aaaaaaa")))
			}
			defer ep.queryPool.Put(pv)

			v, ok := pv.(reflect.Value)
			if !ok {
				panic(wrapErr(fmt.Errorf("queryPool returned value of type not equal to reflect.Value")))
			}
			reflect.ValueOf(rd).Elem().FieldByName("QueryParams").Set(v)

			bs, err := json.Marshal(r.URL.Query())
			if err != nil && err != io.EOF {
				// This block will never run so not tested
				log.Println(wrapErr(err, "json marshal url query failed"))
				badrequest("invalid or missing query params")
				return
			}

			if len(bs) > 0 {
				if err := rd.QueryParams.Unmarshal(bs); err != nil {
					/*
					  This part is not very helpful because it will be valid for any Payload
					  type that inherently has a string->[]string structure and will fail
					  for every other case.
					  Ideally the structure should be verified using reflection.
					  This Unmarshal failing will always indicate the one case above
					*/
					log.Println(wrapErr(err, "query unmarshal failed"))
					badrequest("invalid or missing query params")
					return
				}
				// Test if empty map
				testm := map[string][]string{}
				json.Unmarshal(bs, &testm)
				if len(testm) == 0 {
					badrequest("empty query params")
					return
				}
			} else {
				badrequest("empty query params")
				return
			}
		}

		rc, ok := rcPool.Get().(*RequestCtx)
		if !ok {
			panic(`rcpool returned something thats not a RequestCtx... aaaaaaaaa!!`)
		}
		defer func() {
			rc.Reset()
			rcPool.Put(rc)
		}()
		rc.update(w, r)

		resp, err := ep.handler(rc, rd)
		if err != nil {
			if err := errorHandler(rc, err); err != nil {
				log.Println(wrapErr(err))
			}
			return
		}

		if rc.ResponseWriter.written {
			return
		}

		var resBody []byte
		err = nil
		if resp != nil {
			resBody, err = resp.Marshal()
			if err != nil {
				log.Println(wrapErr(err))
				if err := errorHandler(rc, NewError(StatusInternalServerError)); err != nil {
					log.Println(wrapErr(err))
				}
				return
			}
		}
		rc.ResponseWriter.WriteHeader(StatusOK)
		rc.ResponseWriter.Write(resBody)
	})
}

func (ep *endpoint) pathItem() (*openapi3.PathItem, error) {
	// TODO
	return &openapi3.PathItem{}, nil
}

type EndpointPayload struct {
	RequestPayload  Payload
	QueryPayload    Payload
	ResponsePayload Payload
}

func NewEndpointPayload(ps ...Payload) EndpointPayload {
	var ep EndpointPayload
	if len(ps) > 0 {
		ep.RequestPayload = ps[0]
	}

	if len(ps) > 1 {
		ep.QueryPayload = ps[1]
	}

	if len(ps) > 2 {
		ep.ResponsePayload = ps[2]
	}
	return ep
}

type EndpointConfig struct {
	Route              string
	Handler            Handler
	Payload            EndpointPayload
	ExcludeMiddlewares []string
	method             string
}

func NewEndpointConfig(route string, handler Handler) EndpointConfig {
	return EndpointConfig{
		Route:   route,
		Handler: handler,
	}
}

func (ec EndpointConfig) WithExclude(ms ...string) EndpointConfig {
	ec.ExcludeMiddlewares = append(ec.ExcludeMiddlewares, ms...)
	return ec
}

func (ec EndpointConfig) WithHandler(h Handler) EndpointConfig {
	ec.Handler = h
	return ec
}

func (ec EndpointConfig) WithPayload(ep EndpointPayload) EndpointConfig {
	ec.Payload = ep
	return ec
}

func (ec EndpointConfig) WithRoute(r string) EndpointConfig {
	ec.Route = r
	return ec
}

func (ec *EndpointConfig) applyMiddlerwares(ms []*Middleware) {
	exm := map[string]bool{}
	for _, s := range ec.ExcludeMiddlewares {
		exm[s] = true
	}

	for i := len(ms) - 1; i >= 0; i-- {
		m := ms[i]
		if _, ok := exm[m.ID]; ok {
			continue
		}
		ec.Handler = m.Handler(ec.Handler)
	}
}

func (ec EndpointConfig) endpoint() *endpoint {
	ep := &endpoint{
		method:          ec.method,
		route:           ec.Route,
		handler:         ec.Handler,
		requestPayload:  ec.Payload.RequestPayload,
		queryPayload:    ec.Payload.QueryPayload,
		responsePayload: ec.Payload.ResponsePayload,
		mexclusions:     ec.ExcludeMiddlewares,
	}
	ep.initPools()
	return ep
}
