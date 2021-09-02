package gate

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/julienschmidt/httprouter"
)

type endpoint struct {
	route   string
	handler Handler
	Payload Payload
	rw      http.ResponseWriter
	r       *http.Request
	params  httprouter.Params
	typ     reflect.Type
	val     reflect.Value
}

var epPool sync.Pool

func init() {
	epPool.New = func() interface{} {
		return new(endpoint)
	}
}

func (ep *endpoint) update(
	route string, pl Payload, h Handler,
	rw http.ResponseWriter, r *http.Request, p httprouter.Params,
) {
	ep.route = route
	ep.handler = h
	ep.rw = rw
	ep.r = r
	ep.params = p

	if pl == nil {
		return
	}

	ep.typ = reflect.TypeOf(pl).Elem()
	ep.val = reflect.ValueOf(pl)
}

func (ep *endpoint) reset() {
	ep.handler = nil
	ep.Payload = nil
	ep.rw = nil
	ep.r = nil
	ep.params = nil
	ep.route = ""
	ep.typ = nil
	ep.val = reflect.Value{}
}

func (ep *endpoint) handle() {
	if ep.r.Method == http.MethodPost ||
		ep.r.Method == http.MethodPut ||
		ep.r.Method == http.MethodDelete ||
		ep.r.Method == http.MethodPatch {
		bs, err := io.ReadAll(ep.r.Body)
		if err != nil {
			if err != io.EOF {
				log.Println(`[WARN] -> Payload is not empty. Error reading request body`, err.Error())
			}
		} else {
			defer ep.r.Body.Close()
			if len(bs) > 0 && ep.typ != nil {
				pfv := reflect.ValueOf(ep).Elem().FieldByName("Payload")
				if ep.val.Kind() == reflect.Ptr {
					var v reflect.Value
					if pool, ok := ppm[ep.route]; ok {
						pv := pool.Get()
						if pv == nil {
							panic(fmt.Errorf("route: %s pool returned nil... aaaaaaa", ep.route))
						}
						defer pool.Put(pv)
						if v, ok = pv.(reflect.Value); !ok {
							panic(fmt.Errorf("route: %s pool returned value thats not reflect.Value... aaaaaaa", ep.route))
						}
					} else {
						j := reflect.New(ep.typ).Elem()
						j.Set(ep.val.Elem())
						v = j.Addr()
					}
					pfv.Set(v)
				} else {
					pfv.Set(ep.val)
				}

				// log.Println("Payload: ", ep.Payload)
				if err := ep.Payload.Unmarshal(bs); err != nil {
					log.Println("[ERR] -> Unmarshal to Payload failed", err.Error())
					return
				}
			}
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

	rc.update(ep.rw, ep.r, ep.params)

	resp, err := ep.handler(rc, ep.Payload)
	if err != nil {
		if err := errorHandler(rc, err); err != nil {
			log.Println("[ERR] -> responding err")
			return
		}
	}
	if resp != nil {
		bs, err := resp.Marshal()
		if err != nil {
			log.Println("[ERR] -> [response] -> Payload.Marshal -> ", err.Error())
			if err := errorHandler(rc, ErrInternalServerError); err != nil {
				log.Println("[ERR] -> err in responding with auto generated Internal Server Error", err.Error())
			}
			return
		}
		ep.rw.WriteHeader(StatusOK)
		ep.rw.Write(bs)
		return
	}

	ep.rw.WriteHeader(StatusOK)
}
