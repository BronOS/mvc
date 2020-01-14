package mvc

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"time"
)

// HTTPServerConfig a Server defines parameters for running an HTTP server.
// The zero value for Server is a valid configuration.
type HTTPServerConfig struct {
	Addr string // TCP address to listen on, ":http" if empty

	// TLSConfig optionally provides a TLS configuration for use
	// by ServeTLS and ListenAndServeTLS. Note that this value is
	// cloned by ServeTLS and ListenAndServeTLS, so it's not
	// possible to modify the configuration with methods like
	// tls.Config.SetSessionTicketKeys. To use
	// SetSessionTicketKeys, use Server.Serve with a TLS Listener
	// instead.
	TLSConfig *tls.Config

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int

	// TLSNextProto optionally specifies a function to take over
	// ownership of the provided TLS connection when an NPN/ALPN
	// protocol upgrade has occurred. The map key is the protocol
	// name negotiated. The Handler argument should be used to
	// handle HTTP requests and will initialize the Request's TLS
	// and RemoteAddr if not already set. The connection is
	// automatically closed when the function returns.
	// If TLSNextProto is not nil, HTTP/2 support is not enabled
	// automatically.
	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)

	// ConnState specifies an optional callback function that is
	// called when a client connection changes state. See the
	// ConnState type and associated constants for details.
	ConnState func(net.Conn, http.ConnState)

	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger

	// BaseContext optionally specifies a function that returns
	// the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's
	// about to start accepting requests.
	// If BaseContext is nil, the default is context.Background().
	// If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection c. The provided ctx
	// is derived from the base context and has a ServerContextKey
	// value.
	ConnContext func(ctx context.Context, c net.Conn) context.Context
}

// HTTPServerInterface a http server interface based on mux.Router and http.Server.
type HTTPServerInterface interface {
	// GetServer getter for http.Server
	GetServer() *http.Server

	// GetRouter getter for mux.Router
	GetRouter() *mux.Router

	// AddMiddleware appends a middleware to the chain.
	// Middleware can be used to intercept or otherwise modify requests and/or responses,
	// and are executed in the order that they are applied to the Router.
	AddMiddleware(middleware MiddlewareInterface)

	// AddRoute register controller with a matcher for the URL path and http methods.
	AddRoute(route string, controller ControllerInterface, methods ...string)

	// Run listens on the TCP network address.
	// Always returns a non-nil error. After Shutdown or Close,
	// the returned error is ErrServerClosed.
	Run() error

	// Shutdown gracefully shuts down the server without interrupting any
	// active connections. Shutdown works by first closing all open
	// listeners, then closing all idle connections, and then waiting
	// indefinitely for connections to return to idle and then shut down.
	// If the provided context expires before the shutdown is complete,
	// Shutdown returns the context's error, otherwise it returns any
	// error returned from closing the Server's underlying Listener(s).
	//
	// When Shutdown is called, Serve, ListenAndServe, and
	// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
	// program doesn't exit and waits instead for Shutdown to return.
	//
	// Shutdown does not attempt to close nor wait for hijacked
	// connections such as WebSockets. The caller of Shutdown should
	// separately notify such long-lived connections of shutdown and wait
	// for them to close, if desired. See RegisterOnShutdown for a way to
	// register shutdown notification functions.
	//
	// Once Shutdown has been called on a server, it may not be reused;
	// future calls to methods such as Serve will return ErrServerClosed.
	Shutdown(ctx context.Context) error
}

// HTTPServer a http server based on mux.Router and http.Server.
type HTTPServer struct {
	server *http.Server
	router *mux.Router
}

// NewHTTPServer creates new http server based on passed config.
func NewHTTPServer(config *HTTPServerConfig) *HTTPServer {
	router := mux.NewRouter()
	return &HTTPServer{
		server: &http.Server{
			Addr:              config.Addr,
			Handler:           router,
			TLSConfig:         config.TLSConfig,
			ReadTimeout:       config.ReadTimeout,
			ReadHeaderTimeout: config.ReadHeaderTimeout,
			WriteTimeout:      config.WriteTimeout,
			IdleTimeout:       config.IdleTimeout,
			MaxHeaderBytes:    config.MaxHeaderBytes,
			TLSNextProto:      config.TLSNextProto,
			ConnState:         config.ConnState,
			ErrorLog:          config.ErrorLog,
			BaseContext:       config.BaseContext,
			ConnContext:       config.ConnContext,
		},
		router: router,
	}
}

// GetServer getter for http.Server
func (s *HTTPServer) GetServer() *http.Server {
	return s.server
}

// GetRouter getter for mux.Router
func (s *HTTPServer) GetRouter() *mux.Router {
	return s.router
}

// AddMiddleware appends a middleware to the chain.
// Middleware can be used to intercept or otherwise modify requests and/or responses,
// and are executed in the order that they are applied to the Router.
func (s *HTTPServer) AddMiddleware(middleware MiddlewareInterface) {
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			herr := middleware.Handle(w, r)
			if herr == nil {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, herr.Err.Error(), herr.ResponseCode)
			}
		})
	})
}

// AddRoute register controller with a matcher for the URL path and http methods.
func (s *HTTPServer) AddRoute(route string, controller ControllerInterface, methods ...string) {
	s.router.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		herr := controller.Action(w, r)
		if herr != nil {
			http.Error(w, herr.Err.Error(), herr.ResponseCode)
		}
	}).Methods(methods...)
}

// Run listens on the TCP network address.
// Always returns a non-nil error. After Shutdown or Close,
// the returned error is ErrServerClosed.
func (s *HTTPServer) Run() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and
// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. See RegisterOnShutdown for a way to
// register shutdown notification functions.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
