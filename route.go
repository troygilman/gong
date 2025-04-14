package gong

import (
	"context"
	"fmt"
	"io"
)

// Route represents a route in the application's routing tree.
// It defines the interface for handling component routing and rendering.
type Route interface {
	// Child returns the child route for the given path.
	// If no exact match is found and a default child exists, returns the default child.
	// Returns nil if no matching route is found.
	Child(path string) Route

	// Children returns all direct child routes of this route.
	Children() []Route

	Parent() Route

	// Root returns the root route of the routing tree.
	Root() Route

	// Path returns the path segment that this route represents.
	Path() string

	// Component returns the component associated with this route.
	Component() Component

	// Render renders the route's component to the given writer.
	// If the context indicates an action is being performed, it will render
	// the specific component identified by the context's ID.
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

	if gCtx.link {
		parent := route.Parent()
		if parent == nil {
			return nil
		}
		gCtx.route = parent
		gCtx.link = false
		if component, ok := parent.Component().Find(gCtx.id); ok {
			gCtx.action = true
			if err := render(ctx, gCtx, w, component); err != nil {
				return err
			}
			gCtx.action = false
		}
		return render(ctx, gCtx, w, NewOutlet().withRoute(route).withOOB(true))
	}

	gCtx.route = route

	if gCtx.action {
		component, ok := route.component.Find(gCtx.id)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s", gCtx.id))
		}
		return render(ctx, gCtx, w, component)
	}

	gCtx.id = ""
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

func (route *gongRoute) Parent() Route {
	return route.parent
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
