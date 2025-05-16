package server

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/response_writer"
	"github.com/troygilman/gong/internal/util"
	"github.com/troygilman/gong/route"
)

// Option is a function type for configuring servers with the options pattern.
// It takes a Server pointer and returns a modified Server pointer.
type Option func(*Server) *Server

// WithErrorHandler sets a custom error handler for the server.
// The handler will be called when errors occur during request processing.
func WithErrorHandler(handler gong.ErrorHandler) Option {
	return func(s *Server) *Server {
		s.errorHandler = handler
		return s
	}
}

// Server is the main framework instance that handles routing and request processing.
// It implements the http.Handler interface and manages the application's routes.
type Server struct {
	mux          *http.ServeMux
	routes       []gong.Route
	errorHandler gong.ErrorHandler
}

// New creates a new Server instance.
// It accepts optional configurations via the Option pattern.
func New(opts ...Option) *Server {
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
func (svr *Server) Route(route gong.Route) {
	svr.routes = append(svr.routes, route)
}

// Run starts the server and begins listening for HTTP requests on the specified address.
// This method blocks until the server is stopped or encounters an error.
func (svr *Server) Run(addr string) error {
	root := route.New("", component.New(indexComponent{}), route.WithChildren(svr.routes...))

	for i := range root.NumChildren() {
		child := root.Child(i)
		svr.setupRoute(root, root, child, strconv.Itoa(i), child.Path(), 1)
	}

	return http.ListenAndServe(addr, svr.mux)
}

func (svr *Server) setupRoute(
	root gong.Route,
	parent gong.Route,
	route gong.Route,
	routeID string,
	path string,
	depth int,
) {
	log.Printf("Route=%s\n", path)

	svr.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			writer      = response_writer.NewResponseWriter(w)
			requestType = r.Header.Get(gong.HeaderGongRequestType)
		)

		gCtx := gctx.Context{
			Request:      r,
			Writer:       writer,
			Action:       requestType == gong.GongRequestTypeAction,
			Link:         requestType == gong.GongRequestTypeLink,
			ComponentID:  r.Header.Get(gong.HeaderGongComponentID),
			ErrorHandler: svr.errorHandler,
		}

		switch requestType {
		case gong.GongRequestTypeAction:
			gCtx.RequestRouteID = routeID
			gCtx.CurrentRouteID = r.Header.Get(gong.HeaderGongRouteID)
			gCtx.Route, gCtx.Depth = root.Find(gCtx.CurrentRouteID)
		case gong.GongRequestTypeLink:
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

		if err := util.Render(r.Context(), gCtx, writer, gCtx.Route); err != nil {
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
