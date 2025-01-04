package gong

import (
	"context"
	"log"
	"net/http"
	"reflect"

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

func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Gong) Route(path string, view View, f func(Route)) {
	route := &route{
		gong:     g,
		path:     path,
		view:     view,
		actions:  make(map[string]Action),
		children: make(map[string]*route),
	}
	g.handleRoute(route)
	if f != nil {
		f(route)
	}
}

func (g *Gong) handleRoute(route *route) {
	scanViewForActions(route.actions, route.view, "")
	log.Printf("Route=%s Actions=%#v\n", route.path, route.actions)

	g.handle(route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestType := r.Header.Get(GongRequestHeader)

		gCtx := gongContext{
			requestType: requestType,
			route:       route,
			path:        r.Header.Get(GongRouteHeader),
			request:     r,
			action:      requestType == GongRequestTypeAction,
			kind:        r.Header.Get(GongKindHeader),
		}

		var component templ.Component
		switch requestType {
		case GongRequestTypeAction:
			gCtx.route = route.getRoute(gCtx.path)
			component = gCtx.route
		case GongRequestTypeRoute:
			gCtx.kind = ""
			component = gCtx.route
		default:
			gCtx.path = route.path
			gCtx.route = route.getRoot()
			component = index(gCtx.route)
		}

		if err := render(r.Context(), gCtx, w, component); err != nil {
			panic(err)
		}
	}))
}

func scanViewForActions(actions map[string]Action, view View, kindPrefix string) {
	v := reflect.ValueOf(view)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			kind, ok := t.Field(i).Tag.Lookup("kind")
			if !ok {
				continue
			}
			kind = kindPrefix + kind
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			if action, ok := field.Interface().(Action); ok {
				actions[kind] = action
			}
			if view, ok := field.Interface().(View); ok {
				scanViewForActions(actions, view, kind+"_")
			}
		}
	}
}

func (g *Gong) handle(path string, handler http.Handler) {
	g.mux.Handle(path, handler)
}

type gongContext struct {
	requestType string
	route       *route
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
