package gong

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/troygilman/gong/internal/response_writer"
)

// Option is a function type for configuring servers with the options pattern.
// It takes a Server pointer and returns a modified Server pointer.
type ServerOption func(*Server) *Server

// WithErrorHandler sets a custom error handler for the server.
// The handler will be called when errors occur during request processing.
func WithErrorHandler(handler ErrorHandler) ServerOption {
	return func(s *Server) *Server {
		s.errorHandler = handler
		return s
	}
}

// Server is the main framework instance that handles routing and request processing.
// It implements the http.Handler interface and manages the application's routes.
type Server struct {
	mux          *http.ServeMux
	routes       []Route
	errorHandler ErrorHandler
}

// New creates a new Server instance.
// It accepts optional configurations via the Option pattern.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	for _, opt := range opts {
		s = opt(s)
	}
	return s
}

// Handle registers a handler for the given pattern in the server's HTTP mux.
func (svr *Server) Handle(pattern string, handler http.Handler) {
	svr.mux.Handle(pattern, handler)
}

// Route registers a route with the server.
// The route will be set up with appropriate handlers when the server runs.
func (svr *Server) Route(route Route) {
	svr.routes = append(svr.routes, route)
}

// Run starts the server and begins listening for HTTP requests on the specified address.
// This method blocks until the server is stopped or encounters an error.
func (svr *Server) Run(addr string) error {
	root := NewRoute("", NewComponent(indexComponent{}), WithChildren(svr.routes...))

	for i := range root.NumChildren() {
		child := root.Child(i)
		svr.setupRoute(root, root, child, strconv.Itoa(i), child.Path(), 1)
	}

	return http.ListenAndServe(addr, svr.mux)
}

func (svr *Server) setupRoute(
	root Route,
	parent Route,
	route Route,
	routeID string,
	path string,
	depth int,
) {
	log.Printf("Route=%s\n", path)

	svr.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			writer      = response_writer.NewResponseWriter(w)
			requestType = r.Header.Get(HeaderGongRequestType)
		)

		gCtx := gongContext{
			Request:      r,
			Writer:       writer,
			Action:       requestType == GongRequestTypeAction,
			Link:         requestType == GongRequestTypeLink,
			ComponentID:  r.Header.Get(HeaderGongComponentID),
			ErrorHandler: svr.errorHandler,
		}

		switch requestType {
		case GongRequestTypeAction:
			gCtx.RequestRouteID = routeID
			gCtx.CurrentRouteID = r.Header.Get(HeaderGongRouteID)
			gCtx.Route, gCtx.Depth = root.Find(gCtx.CurrentRouteID)
		case GongRequestTypeLink:
			currentUrl, err := getCurrentUrl(r)
			if err != nil {
				panic(err)
			}
			if currentUrl.EscapedPath() == r.URL.EscapedPath() {
				w.Header().Set("Hx-Reswap", "none")
				return
			}
			gCtx.RequestRouteID = routeID
			gCtx.CurrentRouteID = routeID[:len(routeID)-1]
			gCtx.Depth = depth - 1
			gCtx.Route = parent
		default:
			gCtx.RequestRouteID = routeID
			gCtx.CurrentRouteID = ""
			gCtx.Depth = 0
			gCtx.Route = root
		}

		// log.Println("RequestPath:", r.URL.Path, "RequestRouteID:", gCtx.RequestRouteID, "CurrentRouteID", gCtx.CurrentRouteID)

		if gCtx.Route == nil {
			panic("route is nil")
		}

		if err := render(r.Context(), gCtx, writer, gCtx.Route); err != nil {
			panic(err)
		}

		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}))

	for i := range route.NumChildren() {
		child := route.Child(i)
		svr.setupRoute(root, route, child, routeID+strconv.Itoa(i), path+child.Path(), depth+1)
	}
}

func getCurrentUrl(r *http.Request) (*url.URL, error) {
	return url.Parse(r.Header.Get("Hx-Current-Url"))
}
