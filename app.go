package gate

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/julienschmidt/httprouter"
)

type App struct {
	http.Server
	router *httprouter.Router
	paths  openapi3.Paths
	Info   *openapi3.Info
}

type AppPanicHandler func(http.ResponseWriter, *http.Request, interface{})

type AppOptions struct {
	Info              openapi3.Info
	Addr              string
	TLSConfig         *tls.Config
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	TLSNextProto      map[string]func(*App, *tls.Conn)
	ConnState         func(net.Conn, http.ConnState)
	ErrorLog          *log.Logger
	BaseContext       func(net.Listener) context.Context
	ConnContext       func(ctx context.Context, c net.Conn) context.Context
}

func (ao AppOptions) server() *http.Server {
	server := http.Server{
		Addr: ":6666",
	}
	if ao.Addr != "" {
		server.Addr = ao.Addr
	}
	if ao.TLSConfig != nil {
		server.TLSConfig = ao.TLSConfig
	}

	if int64(ao.ReadTimeout) > 0 {
		server.ReadTimeout = ao.ReadTimeout
	}

	if int64(ao.ReadHeaderTimeout) > 0 {
		server.ReadHeaderTimeout = ao.ReadHeaderTimeout
	}

	if int64(ao.WriteTimeout) > 0 {
		server.WriteTimeout = ao.WriteTimeout
	}

	if int64(ao.IdleTimeout) > 0 {
		server.IdleTimeout = ao.IdleTimeout
	}

	if ao.MaxHeaderBytes > 0 {
		server.MaxHeaderBytes = ao.MaxHeaderBytes
	}

	if len(ao.TLSNextProto) > 0 {
		v := map[string]func(*http.Server, *tls.Conn, http.Handler){}
		for k, f := range ao.TLSNextProto {
			v[k] = func(s *http.Server, c *tls.Conn, h http.Handler) {
				a := new(App)
				a.FromServer(s)
				a.router = newRouter()
				a.Handler = a.router
				f(a, c)
			}
		}
	}
	return &server
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

func New(ao AppOptions) (*App, error) {
	server := ao.server()
	app := &App{}
	app.router = newRouter()
	app.Handler = app.router
	app.FromServer(server)
	app.UpdateInfo(ao.Info)

	if app.Info.Title == "" {
		return nil, wrapErr(fmt.Errorf("AppOptions.Info.Title cannot be empty"))
	}

	if app.Info.Version == "" {
		return nil, wrapErr(fmt.Errorf("AppOptions.Info.Version cannot be empty"))
	}
	app.paths = openapi3.Paths{}
	return app, nil
}

func (a *App) UpdateInfo(i openapi3.Info) {
	if a.Info == nil {
		a.Info = new(openapi3.Info)
	}

	if i.Extensions != nil {
		a.Info.Extensions = i.Extensions
	}

	if i.Title != "" {
		a.Info.Title = i.Title
	}

	if i.Description != "" {
		a.Info.Description = i.Description
	}

	if i.TermsOfService != "" {
		a.Info.TermsOfService = i.TermsOfService
	}

	if i.Contact != nil {
		a.Info.Contact = i.Contact
	}

	if i.License != nil {
		a.Info.License = i.License
	}

	if i.Version != "" {
		a.Info.Version = i.Version
	}
}

func (a *App) FromServer(server *http.Server) {
	a.Addr = server.Addr
	a.TLSConfig = server.TLSConfig
	a.ReadTimeout = server.ReadTimeout
	a.ReadHeaderTimeout = server.ReadHeaderTimeout
	a.WriteTimeout = server.WriteTimeout
	a.IdleTimeout = server.IdleTimeout
	a.MaxHeaderBytes = server.MaxHeaderBytes
	a.TLSNextProto = server.TLSNextProto
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func errorHandler(rc *RequestCtx, err error) error {
	code := StatusInternalServerError
	if e, ok := err.(*Error); ok {
		code = e.Code
	}
	rc.ResponseWriter.WriteHeader(code)
	if _, err := rc.ResponseWriter.Write([]byte(err.Error())); err != nil {
		return err
	}
	return nil
}

// var responseBodyPool = map[string]sync.Pool{}

func (app *App) registerEndpoint(
	method, route string, h Handler,
	f func(string, httprouter.Handle),
	ps ...Payload,
) error {
	var payloadInputs []Payload
	if len(ps) > 0 {
		payloadInputs = append(payloadInputs, ps[0])
	}

	if len(ps) > 1 {
		payloadInputs = append(payloadInputs, ps[1])
	}

	if len(ps) > 2 {
		payloadInputs = append(payloadInputs, ps[2])
	}

	ep := new(endpoint)
	ep.update(method, route, h, payloadInputs...)
	ep.handle(f)
	pi, err := ep.pathItem()
	if err != nil {
		return wrapErr(err)
	}
	app.paths[route] = pi
	return nil
}

type HandleFuncType func(string, Handler, ...Payload) error

// HTTP Method specific handler registrations.
// All route specific restrictions of httprouter apply
// The ps (Payload) variadic input at the end can accept upto 3 objects
// The first one, when exists, determines the request body type
// When provided the RequestData struct received in the handler will have
// request.Body unmarshaled to the Body field.
// request.Body will hence be empty
// The second one, when exists, determines the query parameter type
// When provided the RequestData struct received in the handler will have
// the url query unmarshaled to the QueryParams field
// The third one, when exists, determines the response body type
// It is recommended to provide the above types to generate a valid openapi schema
func (app *App) Get(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodGet, r, h, app.router.GET,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Post(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodPost, r, h, app.router.POST,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Delete(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodDelete, r, h, app.router.DELETE,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Put(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodPut, r, h, app.router.PUT,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Patch(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodPatch, r, h, app.router.PATCH,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Options(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodOptions, r, h, app.router.OPTIONS,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Head(r string, h Handler, ps ...Payload) error {
	if err := app.registerEndpoint(
		http.MethodHead, r, h, app.router.HEAD,
		ps...,
	); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Listen() error {
	if app == nil || app.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	if app.TLSConfig != nil {
		conn, err := net.Listen("tcp", app.Addr)
		if err != nil {
			return wrapErr(err)
		}

		tlsListener := tls.NewListener(conn, app.TLSConfig)
		if err := app.Serve(tlsListener); err != nil {
			return wrapErr(err)
		}
	} else {
		if err := app.ListenAndServe(); err != nil {
			return wrapErr(err)
		}
	}
	return nil
}
