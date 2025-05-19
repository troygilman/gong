package gong

import (
	"log"
	"net/http"
	"net/url"

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
	root := NewRoute("", NewComponent(indexComponent{}), WithChildren(svr.routes...)).newNode(nil, "")
	for _, node := range root.children {
		svr.setupRoute(root, node)
	}
	return http.ListenAndServe(addr, svr.mux)
}

func (svr *Server) setupRoute(root *routeNode, node *routeNode) {
	log.Printf("Route=%s\n", node.path)

	svr.mux.Handle(node.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			writer      = response_writer.NewResponseWriter(w)
			requestType = r.Header.Get(HeaderGongRequestType)
		)

		gCtx := gongContext{
			Request:      r,
			Writer:       writer,
			Action:       requestType == GongRequestTypeAction,
			Link:         requestType == GongRequestTypeLink,
			RouteID:      node.id,
			ComponentID:  r.Header.Get(HeaderGongComponentID),
			RenderedPath: getCurrentUrl(r),
			ErrorHandler: svr.errorHandler,
		}

		switch requestType {
		case GongRequestTypeAction:
			gCtx.Node = root.find(r.Header.Get(HeaderGongRouteID))
		case GongRequestTypeLink:
			gCtx.Node = root
		default:
			gCtx.Node = root
		}

		log.Println("RequestPath:", r.URL.Path, "RouteID:", gCtx.RouteID)

		if gCtx.Node == nil {
			panic("route is nil")
		}

		if err := render(r.Context(), gCtx, writer, gCtx.Node); err != nil {
			panic(err)
		}

		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}))

	for _, child := range node.children {
		svr.setupRoute(root, child)
	}
}

func getCurrentUrl(r *http.Request) string {
	currentUrl := r.Header.Get("Hx-Current-Url")
	u, err := url.Parse(currentUrl)
	if err != nil {
		return ""
	}
	return u.EscapedPath()
}
