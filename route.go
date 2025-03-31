package gong

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

type Route struct {
	path         string
	component    Component
	actions      map[string]Action
	children     map[string]*Route
	defaultChild *Route
	parent       *Route
}

func (route *Route) setupHandler(g *Gong) {
	log.Printf("Route=%s Actions=%#v\n", route.path, route.actions)

	g.handle(route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if err := render(r.Context(), gCtx, writer, component); err != nil {
			panic(err)
		}
		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}))
}

func (r *Route) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.route = r
	gCtx.loader = r.component.loader

	if gCtx.action {
		if action, ok := r.actions[gCtx.kind]; ok {
			gCtx.loader = nil
			if loader, ok := action.(Loader); ok {
				gCtx.loader = loader
			}
			return render(ctx, gCtx, w, action.Action())
		}
		if r.component.action != nil {
			return render(ctx, gCtx, w, r.component.action.Action())
		}
		return nil
	}

	return render(ctx, gCtx, w, r.component.view.View())
}

func (r *Route) getRoute(path string) *Route {
	if r.path == path {
		return r
	}
	return r.children[path].getRoute(path)
}

func (r *Route) getRoot() *Route {
	if r.parent == nil {
		return r
	}
	return r.parent
}
