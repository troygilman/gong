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

func (g *Gong) Handle(path string, handler Handler) {
	log.Printf("registering handler %T on path %s\n", handler, path)

	v := reflect.ValueOf(handler)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			field := t.Field(i)
			subPath, ok := field.Tag.Lookup("path")
			if ok {
				if handler, ok := v.Field(i).Interface().(Handler); ok {
					g.Handle(path+subPath, handler)
				}
			}
		}
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.Header.Get(GongActionHeader)
		ctx := context.WithValue(r.Context(), contextKey, gongContext{
			request: r,
		})
		if action != "" {
			handler = handler.Action(ctx)
		}
		if err := Component(handler, path).Render(ctx, w); err != nil {
			panic(err)
		}
	})
	g.mux.Handle(path, h)
}

type gongContext struct {
	gong            *Gong
	request         *http.Request
	path            string
	action          string
	lastComponentID string
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

type Producer func(ctx context.Context) Handler

type Handler interface {
	Loader(ctx context.Context) Handler
	Action(ctx context.Context) Handler
	Component() templ.Component
}

type component struct {
	handler Handler
	path    string
}

func Component(handler Handler, path string) templ.Component {
	return component{
		handler: handler,
		path:    path,
	}
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.path = gCtx.path + c.path
	ctx = context.WithValue(ctx, contextKey, gCtx)
	handler := c.handler.Loader(ctx)
	return componentWrapper(handler.Component()).Render(ctx, w)
}
