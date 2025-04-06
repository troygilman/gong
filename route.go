package gong

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

type Route interface {
	Child(path string) Route
	Children() []Route
	Root() Route
	Path() string
	Component() Component
	Render(ctx context.Context, w io.Writer) error
}

type gongRoute struct {
	path         string
	component    Component
	actions      map[string]Action
	children     map[string]Route
	defaultChild Route
	parent       Route
}

func (route *gongRoute) setupHandler(mux Mux) {
	log.Printf("Route=%s Actions=%#v\n", route.path, route.actions)

	mux.Handle(route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			gCtx.path = route.path
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
}

func (route *gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.route = route
	gCtx.loader = route.component.loader

	if gCtx.action {
		if action, ok := route.actions[gCtx.kind]; ok {
			gCtx.loader = nil
			if loader, ok := action.(Loader); ok {
				gCtx.loader = loader
			}
			return render(ctx, gCtx, w, action.Action())
		}
		if route.component.action != nil {
			return render(ctx, gCtx, w, route.component.action.Action())
		}
		return nil
	}

	return render(ctx, gCtx, w, route.component.view.View())
}

func (route *gongRoute) Child(path string) Route {
	if child, ok := route.children[path]; ok {
		if child.Path() == path {
			return child
		} else {
			return child.Child(path)
		}
	}
	if route.defaultChild != nil {
		return route.defaultChild
	}
	return nil
}

func (route *gongRoute) Children() []Route {
	children := []Route{}
	for _, route := range route.children {
		children = append(children, route)
	}
	return children
}

func (route *gongRoute) Root() Route {
	if route.parent == nil {
		return route
	}
	return route.parent
}

func (route *gongRoute) Component() Component {
	return route.component
}

func (route *gongRoute) Path() string {
	return route.path
}
