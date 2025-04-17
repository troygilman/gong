package gong

import (
	"context"
	"fmt"
	"io"
	"log"
)

// Route represents a route in the application's routing tree.
// It defines the interface for handling component routing and rendering.
type Route interface {
	// Child returns the child route for the given path.
	// If no exact match is found and a default child exists, returns the default child.
	// Returns nil if no matching route is found.
	Child(int) Route

	// Children returns all direct child routes of this route.
	NumChildren() int

	Parent() Route

	// Root returns the root route of the routing tree.
	Root() Route

	// Path returns the path segment that this route represents.
	Path() string

	FullPath() string

	ID() string

	Depth() int

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
	children     []Route
	defaultChild Route
	parent       Route
	id           string
}

func (route *gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.route = route
	gCtx.path = buildRealPath(route, gCtx.request)

	log.Println("Route:", route.path, route.id)
	if len(route.children) > 0 {
		depth := route.Depth()
		if len(gCtx.routeID) > depth {
			index := int(gCtx.routeID[depth] - '0')
			gCtx.childRoute = route.children[index]
		} else {
			gCtx.childRoute = route.children[0]
		}
	}

	if gCtx.link {
		parent := route.Parent()
		if parent == nil {
			panic("could not find parent")
		}
		gCtx.route = parent
		gCtx.childRoute = route
		gCtx.link = false
		if component, ok := parent.Component().Find(gCtx.componentID); ok {
			gCtx.action = true
			if err := render(ctx, gCtx, w, component); err != nil {
				return err
			}
			gCtx.action = false
		}
		return render(ctx, gCtx, w, NewOutlet().withRoute(route).withOOB(true))
	}

	if gCtx.action {
		component, ok := route.component.Find(gCtx.componentID)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s in route %s", gCtx.componentID, route.path))
		}
		return render(ctx, gCtx, w, component)
	}

	gCtx.componentID = ""
	return render(ctx, gCtx, w, route.component)
}

func (route *gongRoute) Child(index int) Route {
	return route.children[index]
}

func (route *gongRoute) ID() string {
	if route.parent != nil {
		return route.parent.ID() + route.id
	}
	return route.id
}

func (route *gongRoute) NumChildren() int {
	return len(route.children)
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

func (route *gongRoute) FullPath() string {
	if route.parent != nil {
		return route.parent.FullPath() + route.path
	}
	return route.path
}

func (route *gongRoute) Depth() int {
	if route.parent == nil {
		return 0
	} else {
		return route.parent.Depth() + 1
	}
}
