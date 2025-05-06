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

// Gong is the main framework instance that handles routing and request processing.
// It implements the http.Handler interface and manages the application's routes.
type Server struct {
	mux    *http.ServeMux
	routes []gong.Route
}

// New creates a new Gong instance with the specified HTTP mux.
// The mux is used for routing HTTP requests to the appropriate handlers.
func New() *Server {
	return &Server{
		mux: http.NewServeMux(),
	}
}

func (svr *Server) Handle(pattern string, handler http.Handler) {
	svr.mux.Handle(pattern, handler)
}

// Routes registers one or more route builders with the Gong instance.
// Each route builder is built and set up with appropriate handlers.
// Returns the Gong instance for method chaining.
func (svr *Server) Route(route gong.Route) {
	svr.routes = append(svr.routes, route)
}

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
			Request:     r,
			Writer:      writer,
			Action:      requestType == gong.GongRequestTypeAction,
			Link:        requestType == gong.GongRequestTypeLink,
			ComponentID: r.Header.Get(gong.HeaderGongComponentID),
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

		log.Println("RequestPath:", r.URL.Path, "RequestRouteID:", gCtx.RequestRouteID, "CurrentRouteID", gCtx.CurrentRouteID)

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
