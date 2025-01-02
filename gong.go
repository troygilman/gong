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
	HXRequestHeader = "HX-Request"
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

func (g *Gong) Route(route Route) {
	g.route("", NewRoute(route.Path(), Index{
		Handler: route.Handler(),
	}))
}

func (g *Gong) route(path string, route Route) {
	path += route.Path()

	g.handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gCtx := gongContext{
			request: r,
			action:  r.Header.Get(HXRequestHeader) == "true",
		}
		ctx := context.WithValue(r.Context(), contextKey, gCtx)

		component := component{
			route:  route,
			action: r.Header.Get(HXRequestHeader) == "true",
		}

		if err := component.Render(ctx, w); err != nil {
			panic(err)
		}
	}))

	g.decomposeHandler(path, route.Handler())
}

func (g *Gong) handle(path string, handler http.Handler) {
	log.Printf("registering handler %T on path %s\n", handler, path)
	g.mux.Handle(path, handler)
}

func (g *Gong) decomposeHandler(path string, handler Handler) {
	v := reflect.ValueOf(handler)
	if v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			field := v.Field(i)
			if field.CanInterface() {
				switch field := field.Interface().(type) {
				case Route:
					g.route(path, field)
				case Handler:
					g.decomposeHandler(path, field)
				}
			}
		}
	}
}

type gongContext struct {
	request *http.Request
	handler Handler
	path    string
	action  bool
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
	Action() templ.Component
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
	route  Route
	action bool
}

func Component(route Route) templ.Component {
	return component{
		route: route,
	}
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = c.action
	gCtx.handler = c.route.Handler()

	if c.action {
		gCtx.path = gCtx.request.RequestURI
	} else {
		gCtx.path += c.route.Path()
	}

	ctx = context.WithValue(ctx, contextKey, gCtx)

	if gCtx.action {
		if err := target(c.route.Handler().Action()).Render(ctx, w); err != nil {
			return err
		}
		return nil
	}

	return c.route.Handler().Component().Render(ctx, w)
}
