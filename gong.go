package gong

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

type contextKeyType int

const contextKey = contextKeyType(0)

const (
	GongRequestHeader = "Gong-Request"
	GongKindHeader    = "Gong-Kind"
	GongRouteHeader   = "Gong-Route"
)

const (
	GongRequestTypeAction = "action"
	GongRequestTypeRoute  = "route"
)

const (
	TriggerNone = "none"
	TriggerLoad = "load"
)

const (
	SwapNone      = "none"
	SwapOuterHTML = "outerHTML"
	SwapInnerHTML = "innerHTML"
	SwapBeforeEnd = "beforeend"
)

type Gong struct {
	mux Mux
}

func New(mux Mux) *Gong {
	return &Gong{
		mux: mux,
	}
}

func (g *Gong) Routes(builders ...RouteBuilder) *Gong {
	for _, builder := range builders {
		g.setupRoute(builder.build(nil))
	}
	return g
}

func (g *Gong) setupRoute(route Route) {
	log.Printf("Route=%s\n", route.Path())
	g.mux.Handle(route.Path(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := NewCustomResponseWriter(w)
		requestType := r.Header.Get(GongRequestHeader)

		gCtx := gongContext{
			requestType: requestType,
			route:       route,
			path:        r.Header.Get(GongRouteHeader),
			request:     r,
			writer:      writer,
			action:      requestType == GongRequestTypeAction,
			kind:        r.Header.Get(GongKindHeader),
		}

		var templComponent templ.Component
		switch requestType {
		case GongRequestTypeAction:
			if route.Path() != gCtx.path {
				gCtx.route = route.Child(gCtx.path)
			}
			templComponent = gCtx.route
		case GongRequestTypeRoute:
			gCtx.kind = ""
			templComponent = gCtx.route
		default:
			gCtx.path = route.Path()
			gCtx.route = route.Root()
			templComponent = index(gCtx.route)
		}

		if gCtx.route == nil {
			panic("route is nil")
		}

		if err := render(r.Context(), gCtx, writer, templComponent); err != nil {
			panic(err)
		}
		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}))

	for _, child := range route.Children() {
		g.setupRoute(child)
	}
}

func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

type gongContext struct {
	requestType string
	route       Route
	request     *http.Request
	writer      *CustomResponseWriter
	path        string
	action      bool
	loader      Loader
	kind        string
}

type Mux interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Handle(path string, handler http.Handler)
}

type View interface {
	View() templ.Component
}

type Loader interface {
	Loader(ctx context.Context) any
}

type Action interface {
	Action() templ.Component
}

type Head interface {
	Head() templ.Component
}

type LoaderFunc func(ctx context.Context) any

func (f LoaderFunc) Loader(ctx context.Context) any {
	return f(ctx)
}
