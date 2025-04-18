package gong

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/a-h/templ"
)

// Route represents a route in the application's routing tree.
// It defines the interface for handling component routing and rendering.
type Route interface {
	templ.Component

	// Child returns the child route for the given path.
	// If no exact match is found and a default child exists, returns the default child.
	// Returns nil if no matching route is found.
	Child(int) Route

	// NumChildren returns the number of direct child routes of this route.
	NumChildren() int

	// Parent returns the parent route of this route.
	Parent() Route

	// Root returns the root route of the routing tree.
	Root() Route

	// Path returns the path segment that this route represents.
	Path() string

	// FullPath returns the full path of this route, including all parent paths.
	FullPath() string

	// ID returns the unique identifier for this route.
	ID() string

	// Depth returns the depth of this route in the routing tree.
	Depth() int

	// Component returns the component associated with this route.
	Component() Component
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
	componentID := strings.Split(gCtx.componentID, idDelimeter)

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
		if component, ok := parent.Component().Find(componentID); ok {
			if err := render(ctx, gCtx, w, component.Action()); err != nil {
				return err
			}
		}
		return render(ctx, gCtx, w, NewOutlet().withRoute(route).withOOB(true))
	}

	if gCtx.action {
		component, ok := route.component.Find(componentID)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s in route %s", gCtx.componentID, route.path))
		}
		return render(ctx, gCtx, w, component.Action())
	}

	gCtx.componentID = ""
	return render(ctx, gCtx, w, route.component.View())
}

func (route *gongRoute) Child(index int) Route {
	if index < 0 || index >= len(route.children) {
		return nil
	}
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
