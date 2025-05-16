package gong

import (
	"context"
	"fmt"
	"io"
)

// gongRoute is the internal implementation of the gong.Route interface.
// It represents a route in the application's routing tree.
type gongRoute struct {
	path      string
	component Component
	children  []Route
}

// New creates a new Route instance with the specified path and component.
// It accepts optional configurations via the Option pattern.
func NewRoute(path string, component Component, opts ...RouteOption) Route {
	route := gongRoute{
		path:      path,
		component: component,
	}

	for _, opt := range opts {
		route = opt(route)
	}

	return route
}

func (route gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.Route = route
	gCtx.ChildRouteIndex = 0

	// log.Printf("Rendering Route: %+v\n", gCtx)
	if len(route.children) > 0 {
		if len(gCtx.RequestRouteID) > gCtx.Depth {
			gCtx.ChildRouteIndex = int(gCtx.RequestRouteID[gCtx.Depth] - '0')
		}
	}

	if gCtx.Link {
		gCtx.Link = false
		return render(ctx, gCtx, w, NewOutlet(OutletWithOOB(true)))
	}

	if gCtx.Action {
		component, ok := route.component.Find(gCtx.ComponentID)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s in route %s", gCtx.ComponentID, route.path))
		}
		return render(ctx, gCtx, w, component.Action())
	}

	gCtx.ComponentID = ""
	return render(ctx, gCtx, w, route.component.View())
}

// Child returns the child route at the specified index.
// Returns nil if the index is out of bounds.
func (route gongRoute) Child(index int) Route {
	if index < 0 || index >= len(route.children) {
		return nil
	}
	return route.children[index]
}

// Find locates a route by its ID string.
// Returns the found route and the depth in the routing tree.
func (route gongRoute) Find(id string) (Route, int) {
	var r Route = route
	depth := 0
	for _, index := range id {
		r = r.Child(int(index - '0'))
		depth++
	}
	return r, depth
}

// NumChildren returns the number of direct child routes.
func (route gongRoute) NumChildren() int {
	return len(route.children)
}

// Component returns the component associated with this route.
func (route gongRoute) Component() Component {
	return route.component
}

// Path returns the path segment that this route represents.
func (route gongRoute) Path() string {
	return route.path
}

// Option is a function type for configuring routes with the options pattern.
// It takes a gongRoute and returns a modified one.
type RouteOption func(gongRoute) gongRoute

// WithChildren sets the child routes for a route.
// This allows for creating a hierarchical routing structure.
func RouteWithChildren(children ...Route) RouteOption {
	return func(gr gongRoute) gongRoute {
		gr.children = children
		return gr
	}
}
