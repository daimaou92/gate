package gate

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

type App struct {
	router *httprouter.Router
	server *http.Server
}

type AppPanicHandler func(http.ResponseWriter, *http.Request, interface{})

type AppOptions struct {
	Addr                    string
	TLSConfig               *tls.Config
	ReadTimeout             time.Duration
	ReadHeaderTimeout       time.Duration
	WriteTimeout            time.Duration
	IdleTimeout             time.Duration
	MaxHeaderBytes          int
	TLSNextProto            map[string]func(*App, *tls.Conn)
	ConnState               func(net.Conn, http.ConnState)
	ErrorLog                *log.Logger
	BaseContext             func(net.Listener) context.Context
	ConnContext             func(ctx context.Context, c net.Conn) context.Context
	optionsHandler          http.Handler
	methodNotAllowedHandler http.Handler
	panicHandler            AppPanicHandler
}

func (a *App) SetMethodNotAllowedHandler(h http.Handler) error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	a.router.HandleMethodNotAllowed = true
	a.router.MethodNotAllowed = h
	return nil
}

func (a *App) HandleMethodNotAllowed(b bool) error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	a.router.HandleMethodNotAllowed = b
	return nil
}

func (a *App) HandleOPTIONS(b bool) error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	a.router.HandleOPTIONS = b
	return nil
}

func (a *App) SetOptionsHandler(h http.Handler) error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}
	a.router.HandleOPTIONS = true
	a.router.GlobalOPTIONS = h
	return nil
}

func (a *App) SetPanicHandler(h AppPanicHandler) error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}
	a.router.PanicHandler = h
	return nil
}

func newRouter() *httprouter.Router {
	r := httprouter.New()
	r.HandleMethodNotAllowed = true

	r.HandleOPTIONS = true
	return r
}

func New(ao *AppOptions) *App {
	server := &http.Server{}
	if ao != nil {
		server.Addr = ao.Addr
		server.TLSConfig = ao.TLSConfig
		server.ReadTimeout = ao.ReadTimeout
		server.ReadHeaderTimeout = ao.ReadHeaderTimeout
		server.WriteTimeout = ao.WriteTimeout
		server.IdleTimeout = ao.IdleTimeout
		server.MaxHeaderBytes = ao.MaxHeaderBytes
		if len(ao.TLSNextProto) > 0 {
			v := map[string]func(*http.Server, *tls.Conn, http.Handler){}
			for k, f := range ao.TLSNextProto {
				v[k] = func(s *http.Server, c *tls.Conn, h http.Handler) {
					a := New(nil)
					a.server = s
					a.router = newRouter()
					f(a, c)
				}
			}
		}
		server.Handler = newRouter()
	}
	app := App{
		server: server,
	}
	return &app
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func errorHandler(rc *RequestCtx, err error) error {
	code := StatusInternalServerError
	if e, ok := err.(*Error); ok {
		code = e.Code
	}
	rc.rw.WriteHeader(code)
	if _, err := rc.rw.Write([]byte(err.Error())); err != nil {
		return err
	}
	return nil
}

func (a *App) registerEndpoint(
	route string, pl Payload, h Handler,
	rw http.ResponseWriter, r *http.Request, p httprouter.Params,
) {
	ep, ok := epPool.Get().(*endpoint)
	if !ok {
		panic("epPool gave something thats not endpoint... aaaaaaa!!")
	}
	defer func() {
		ep.reset()
		epPool.Put(ep)
	}()
	ep.update(route, pl, h, rw, r, p)
	ep.handle()
}

func (a *App) Get(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodGet, r, pl, h)
}

func (a *App) Post(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodPost, r, pl, h)
}

func (a *App) Delete(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodDelete, r, pl, h)
}

func (a *App) Put(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodPut, r, pl, h)
}

func (a *App) Patch(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodPatch, r, pl, h)
}

func (a *App) Options(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodOptions, r, pl, h)
}

func (a *App) Head(r string, pl Payload, h Handler) {
	registerRoute(a, http.MethodHead, r, pl, h)
}

func (a *App) Listen() error {
	if a == nil || a.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	if a.server.TLSConfig != nil {
		conn, err := net.Listen("tcp", a.server.Addr)
		if err != nil {
			return wrapErr(err)
		}

		tlsListener := tls.NewListener(conn, a.server.TLSConfig)
		if err := a.server.Serve(tlsListener); err != nil {
			return wrapErr(err)
		}
	} else {
		if err := a.server.ListenAndServe(); err != nil {
			return wrapErr(err)
		}
	}
	return nil
}
