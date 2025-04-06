package gong

import (
	"context"
	"fmt"
	"io"
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
	children     map[string]Route
	defaultChild Route
	parent       Route
}

func (route *gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.route = route

	if gCtx.action {
		component, ok := route.component.Find(gCtx.kind)
		if !ok {
			panic(fmt.Sprintf("could not find component with kind %s", gCtx.kind))
		}
		return render(ctx, gCtx, w, component)
	}

	return render(ctx, gCtx, w, route.component)
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
