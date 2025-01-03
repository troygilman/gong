package gong

import (
	"context"
	"io"
	"log"
	"net/http"
	"reflect"

	"github.com/a-h/templ"
)

type contextKeyType int

const contextKey = contextKeyType(0)

const (
	GongActionHeader = "Gong-Action"
	GongKindHeader   = "Gong-Kind"
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

func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Gong) Route(path string, view View, f func(Route)) {
	route := Route{
		gong: g,
		path: path,
		view: Index{
			view: view,
		},
		actions: make(map[string]Action),
	}

	scanViewForActions(route.actions, view)
	g.handleRoute(route)
	f(route)
}

func scanViewForActions(actions map[string]Action, view View) {
	v := reflect.ValueOf(view)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			kind, ok := t.Field(i).Tag.Lookup("kind")
			if !ok {
				continue
			}
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			if action, ok := field.Interface().(Action); ok {
				actions[kind] = action
			}
			if view, ok := field.Interface().(View); ok {
				scanViewForActions(actions, view)
			}
		}
	}
}

func (g *Gong) handleRoute(route Route) {
	g.handle(route.Path(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gCtx := gongContext{
			route:   route,
			path:    route.Path(),
			request: r,
			action:  r.Header.Get(GongActionHeader) == "true",
			kind:    r.Header.Get(GongKindHeader),
		}

		if loader, ok := route.View().(Loader); ok {
			gCtx.loader = loader
		}

		if err := render(r.Context(), gCtx, w, route); err != nil {
			panic(err)
		}
	}))
}

func (g *Gong) handle(path string, handler http.Handler) {
	log.Printf("registering handler %T on path %s\n", handler, path)
	g.mux.Handle(path, handler)
}

type gongContext struct {
	route   Route
	request *http.Request
	path    string
	action  bool
	loader  Loader
	kind    string
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

type LoaderFunc func(ctx context.Context) any

func (f LoaderFunc) Loader(ctx context.Context) any {
	return f(ctx)
}

type component struct {
	kind   string
	view   View
	action bool
	config componentConfig
}

func Component(kind string, view View, opts ...ComponentOption) templ.Component {
	c := component{
		kind: kind,
		view: view,
	}
	if loader, ok := view.(Loader); ok {
		c.config.loader = loader
	}
	for _, opt := range opts {
		c.config = opt(c.config)
	}
	return c
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = c.action
	gCtx.loader = c.config.loader
	gCtx.kind = c.kind
	ctx = context.WithValue(ctx, contextKey, gCtx)

	return c.view.View().Render(ctx, w)
}
