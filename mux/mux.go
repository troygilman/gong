package mux

import (
	"log"
	"net/http"
	"net/url"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/response_writer"
	"github.com/troygilman/gong/internal/util"
	"github.com/troygilman/gong/route"
)

// Gong is the main framework instance that handles routing and request processing.
// It implements the http.Handler interface and manages the application's routes.
type Gong struct {
	mux         *http.ServeMux
	rootBuilder route.Builder
	root        gong.Route
}

// New creates a new Gong instance with the specified HTTP mux.
// The mux is used for routing HTTP requests to the appropriate handlers.
func New() *Gong {
	return &Gong{
		mux:         http.NewServeMux(),
		rootBuilder: route.New("", component.New(indexComponent{})),
	}
}

// Routes registers one or more route builders with the Gong instance.
// Each route builder is built and set up with appropriate handlers.
// Returns the Gong instance for method chaining.
func (g *Gong) Routes(builders ...route.Builder) *Gong {
	g.rootBuilder = g.rootBuilder.WithRoutes(builders...)

	g.root = g.rootBuilder.Build(nil, "")
	for i := range g.root.NumChildren() {
		g.setupRoute(g.root.Child(i))
	}
	return g
}

func (g *Gong) setupRoute(route gong.Route) {
	log.Printf("Route=%s\n", route.FullPath())
	g.mux.Handle(route.FullPath(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			writer      = response_writer.NewResponseWriter(w)
			requestType = r.Header.Get(gong.HeaderGongRequestType)
			routeID     = r.Header.Get(gong.HeaderGongRouteID)
		)

		gCtx := gctx.Context{
			Route:       route,
			RouteID:     route.ID(),
			Request:     r,
			Writer:      writer,
			Action:      requestType == gong.GongRequestTypeAction,
			Link:        requestType == gong.GongRequestTypeLink,
			ComponentID: r.Header.Get(gong.HeaderGongComponentID),
		}

		switch requestType {
		case gong.GongRequestTypeAction:
			gCtx.Route = g.root.Find(routeID)
		case gong.GongRequestTypeLink:
			currentUrl, err := getCurrentUrl(r)
			if err != nil {
				panic(err)
			}
			if currentUrl.EscapedPath() == r.URL.EscapedPath() {
				w.Header().Set("Hx-Reswap", "none")
				return
			}
		default:
			gCtx.Route = g.root
		}

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
		g.setupRoute(route.Child(i))
	}
}

func (g *Gong) Handle(pattern string, handler http.Handler) {
	g.mux.Handle(pattern, handler)
}

// ServeHTTP implements the http.Handler interface.
// It delegates request handling to the underlying mux.
func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func getCurrentUrl(r *http.Request) (*url.URL, error) {
	return url.Parse(r.Header.Get("Hx-Current-Url"))
}
