package gong

import (
	"context"
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
		builder.build(g, nil)
	}
	return g
}

func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Gong) handle(path string, handler http.Handler) {
	g.mux.Handle(path, handler)
}

type gongContext struct {
	requestType string
	route       *Route
	request     *http.Request
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
