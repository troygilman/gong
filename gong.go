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

func (g *Gong) Route(route Route) {
	route = NewRoute(route.Path(), Index{
		Handler: route.Handler(),
	})
	g.route("", route)
}

func (g *Gong) route(path string, route Route) {
	path += route.Path()
	log.Printf("registering handler %T on path %s\n", route.Handler(), path)
	g.handler(path, route.Handler())

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gCtx := gongContext{
			request: r,
			action:  r.Header.Get("HX-Request") == "true",
		}
		ctx := context.WithValue(r.Context(), contextKey, gCtx)
		if gCtx.action {
			route = NewRoute(route.Path(), route.Handler().Action(ctx))
		}
		if err := Component(route).Render(ctx, w); err != nil {
			panic(err)
		}
	})
	g.mux.Handle(path, h)
}

func (g *Gong) handler(path string, handler Handler) {
	v := reflect.ValueOf(handler)
	if v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			field := v.Field(i)
			if field.CanInterface() {
				switch field := field.Interface().(type) {
				case Route:
					g.route(path, field)
				case Handler:
					g.handler(path, field)
				}
			}
		}
	}
}

type gongContext struct {
	request *http.Request
	action  bool
	path    string
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
	Loader(ctx context.Context) Handler
	Action(ctx context.Context) Handler
	Component() templ.Component
}

type Route interface {
	Path() string
	Handler() Handler
}

type route struct {
	path    string
	handler Handler
}

func NewRoute(path string, handler Handler) Route {
	return route{
		path:    path,
		handler: handler,
	}
}

func (r route) Path() string {
	return r.path
}

func (r route) Handler() Handler {
	return r.handler
}

type component struct {
	route Route
}

func Component(route Route) templ.Component {
	return component{
		route: route,
	}
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	if gCtx.action {
		gCtx.path = gCtx.request.RequestURI
		gCtx.action = false
	} else {
		gCtx.path += c.route.Path()
	}
	ctx = context.WithValue(ctx, contextKey, gCtx)
	handler := c.route.Handler().Loader(ctx)
	return componentWrapper(handler.Component()).Render(ctx, w)
}
