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

type epInit struct {
	ec EndpointConfig
	f  func(string, httprouter.Handle)
}
type App struct {
	http.Server
	router      *httprouter.Router
	paths       openapi3.Paths
	Info        *openapi3.Info
	middlewares []*Middleware
	mwareIndex  map[string]int
	epCache     []epInit
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
	app.mwareIndex = map[string]int{}
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

// Called before listen
func (app *App) mountEndpoints() {
	for _, v := range app.epCache {
		v.ec.applyMiddlerwares(app.middlewares)
		ep := v.ec.endpoint()
		ep.handle(v.f)
	}
}

func (app *App) registerEndpoint(
	ec EndpointConfig,
	f func(string, httprouter.Handle),
) {
	app.epCache = append(app.epCache, epInit{
		ec: ec,
		f:  f,
	})
	// ep := ec.endpoint()
	// ep.update(method, route, h, payloadInputs...)
	// ep.handle(f)
	// pi, err := ep.pathItem()
	// if err != nil {
	// 	return wrapErr(err)
	// }
	// app.paths[ec.Route] = pi
}

type HandleFuncType func(EndpointConfig)

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
func (app *App) Get(ec EndpointConfig) {
	ec.method = http.MethodGet
	app.registerEndpoint(
		ec, app.router.GET,
	)
}

func (app *App) Post(ec EndpointConfig) {
	ec.method = http.MethodPost
	app.registerEndpoint(
		ec, app.router.POST,
	)
}

func (app *App) Delete(ec EndpointConfig) {
	ec.method = http.MethodDelete
	app.registerEndpoint(
		ec, app.router.DELETE,
	)
}

func (app *App) Put(ec EndpointConfig) {
	ec.method = http.MethodPut
	app.registerEndpoint(
		ec, app.router.PUT,
	)
}

func (app *App) Patch(ec EndpointConfig) {
	ec.method = http.MethodPatch
	app.registerEndpoint(
		ec, app.router.PATCH,
	)
}

func (app *App) Options(ec EndpointConfig) {
	ec.method = http.MethodOptions
	app.registerEndpoint(
		ec, app.router.OPTIONS,
	)
}

func (app *App) Head(ec EndpointConfig) {
	ec.method = http.MethodHead
	app.registerEndpoint(
		ec, app.router.HEAD,
	)
}

func (app *App) addMiddleware(m *Middleware) error {
	if !m.valid(app) {
		return wrapErr(fmt.Errorf("invalid middleware"))
	}
	app.middlewares = append(app.middlewares, m)
	if app.mwareIndex == nil {
		app.mwareIndex = map[string]int{}
	}
	app.mwareIndex[m.ID] = len(app.middlewares) - 1
	return nil
}

func (app *App) Apply(m *Middleware) error {
	if err := app.addMiddleware(m); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (app *App) Listen() error {
	if app == nil || app.router == nil {
		return wrapErr(fmt.Errorf("app not initialized"))
	}

	app.mountEndpoints()
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
