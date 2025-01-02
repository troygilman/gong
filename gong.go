package gong

import (
	"context"
	"encoding/json"
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

func (g *Gong) Route(path string, handler Handler, f func(Route)) {
	route := Route{
		gong: g,
		path: path,
		handler: Index{
			handler: handler,
		},
		componentActions: make(map[string]Action),
	}

	v := reflect.ValueOf(handler)
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
			action, ok := field.Interface().(Action)
			if !ok {
				continue
			}
			route.componentActions[kind] = action
		}
	}

	g.handleRoute(route)
	f(route)
}

func (g *Gong) handleRoute(route Route) {
	g.handle(route.Path(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gCtx := gongContext{
			route:   route,
			request: r,
			kind:    r.Header.Get(GongKindHeader),
		}

		if loader, ok := route.Handler().(Loader); ok {
			gCtx.loader = loader
		}

		ctx := context.WithValue(r.Context(), contextKey, gCtx)

		component := routeComponent{
			route:  route,
			action: r.Header.Get(GongActionHeader) == "true",
		}

		if err := component.Render(ctx, w); err != nil {
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

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
}

func Bind(ctx context.Context, dest any) error {
	gCtx := getContext(ctx)
	if err := json.NewDecoder(gCtx.request.Body).Decode(dest); err != nil {
		return err
	}
	return nil
}

func Param(ctx context.Context, key string) string {
	gCtx := getContext(ctx)
	return gCtx.request.FormValue(key)
}

type Mux interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Handle(path string, handler http.Handler)
}

type Handler interface {
	Component() templ.Component
}

type Loader interface {
	Loader(ctx context.Context) any
}

type Action interface {
	Action() templ.Component
}

type Route struct {
	gong             *Gong
	path             string
	handler          Handler
	componentActions map[string]Action
}

func (r Route) Route(path string, handler Handler, f func(r Route)) {
	r.path += path
	r.handler = handler
	r.gong.handleRoute(r)
	f(Route{
		gong: r.gong,
		path: r.path,
	})
}

func (r Route) Path() string {
	return r.path
}

func (r Route) Handler() Handler {
	return r.handler
}

type LoaderFunc func(ctx context.Context) any

func (f LoaderFunc) Loader(ctx context.Context) any {
	return f(ctx)
}

type ComponentOption func(component) component

func WithComponentLoader(loader Loader) ComponentOption {
	return func(c component) component {
		c.loader = loader
		return c
	}
}

func WithComponentData(data any) ComponentOption {
	return func(c component) component {
		c.loader = LoaderFunc(func(ctx context.Context) any {
			return data
		})
		return c
	}
}

type component struct {
	kind    string
	handler Handler
	loader  Loader
	action  bool
}

func Component(kind string, handler Handler, opts ...ComponentOption) templ.Component {
	c := component{
		kind:    kind,
		handler: handler,
	}
	if loader, ok := handler.(Loader); ok {
		c.loader = loader
	}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = c.action
	gCtx.loader = c.loader
	gCtx.kind = c.kind
	ctx = context.WithValue(ctx, contextKey, gCtx)

	return c.handler.Component().Render(ctx, w)
}

type routeComponent struct {
	route  Route
	action bool
}

func (rc routeComponent) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = rc.action
	gCtx.path = rc.route.Path()
	ctx = context.WithValue(ctx, contextKey, gCtx)

	if gCtx.action {
		if action, ok := rc.route.componentActions[gCtx.kind]; ok {
			return action.Action().Render(ctx, w)
		}
		if action, ok := rc.route.Handler().(Action); ok {
			return action.Action().Render(ctx, w)
		}
		return nil
	}

	return rc.route.Handler().Component().Render(ctx, w)
}
