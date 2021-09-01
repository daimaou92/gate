package gate

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/julienschmidt/httprouter"
)

var ppm = map[string]sync.Pool{} // Payload Pool Map

func registerRoute(app *App, method, route string, pl Payload, h Handler) {
	type __hf func(string, httprouter.Handle)
	var f __hf

	switch method {
	case http.MethodGet:
		f = app.router.GET
	case http.MethodPost:
		f = app.router.POST
	case http.MethodDelete:
		f = app.router.DELETE
	case http.MethodHead:
		f = app.router.HEAD
	case http.MethodOptions:
		f = app.router.OPTIONS
	case http.MethodPatch:
		f = app.router.PATCH
	case http.MethodPut:
		f = app.router.PUT
	case http.MethodConnect, http.MethodTrace:
		panic("oops!! httprouter does not support these")
	default:
		panic(fmt.Sprintln("the hell kinda method is: ", method))
	}

	if _, ok := ppm[route]; !ok && pl != nil && reflect.TypeOf(pl).Kind() == reflect.Ptr {
		ppm[route] = sync.Pool{
			New: func() interface{} {
				inpt := reflect.TypeOf(pl)
				switch inpt.Kind() {
				case reflect.Array, reflect.Chan,
					reflect.Map, reflect.Ptr, reflect.Slice:
					inpt = inpt.Elem()
				}
				val := reflect.ValueOf(pl)
				v := reflect.New(inpt).Elem()
				v.Set(val.Elem())
				return v.Addr()
			},
		}
	}

	f(route, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		app.registerEndpoint(route, pl, h, rw, r, params)
	})
}
