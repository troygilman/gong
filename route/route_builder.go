package route

import (
	"github.com/troygilman/gong"
)

// RouteBuilder is a builder pattern implementation for creating routes in the Gong framework.
// It allows for declarative route definition with nested child routes.
type Builder struct {
	path      string
	component gong.Component
	children  []Builder
}

// NewRoute creates a new RouteBuilder with the specified path and component.
// The path should be a valid URL path segment, and the component will be rendered
// when this route is matched.
func New(path string, component gong.Component) Builder {
	return Builder{
		path:      path,
		component: component,
	}
}

// WithRoutes adds child routes to the current route builder.
// This allows for creating nested route hierarchies where child routes
// are rendered within their parent route's component.
func (builder Builder) WithRoutes(routes ...Builder) Builder {
	builder.children = routes
	return builder
}

func (builder Builder) Build(parent gong.Route) gong.Route {
	route := &gongRoute{
		component: builder.component,
		path:      builder.path,
		parent:    parent,
	}

	for i, childBuilder := range builder.children {
		child := childBuilder.Build(route)
		route.children = append(route.children, child)
		if i == 0 {
			route.defaultChild = child
		}
	}

	return route
}
