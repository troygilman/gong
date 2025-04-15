package gong

import "strconv"

// RouteBuilder is a builder pattern implementation for creating routes in the Gong framework.
// It allows for declarative route definition with nested child routes.
type RouteBuilder struct {
	path      string
	component Component
	children  []RouteBuilder
}

// NewRoute creates a new RouteBuilder with the specified path and component.
// The path should be a valid URL path segment, and the component will be rendered
// when this route is matched.
func NewRoute(path string, component Component) RouteBuilder {
	return RouteBuilder{
		path:      path,
		component: component,
	}
}

// WithRoutes adds child routes to the current route builder.
// This allows for creating nested route hierarchies where child routes
// are rendered within their parent route's component.
func (builder RouteBuilder) WithRoutes(routes ...RouteBuilder) RouteBuilder {
	builder.children = routes
	return builder
}

func (builder RouteBuilder) build(parent Route, id string) Route {
	route := &gongRoute{
		component: builder.component,
		path:      builder.path,
		parent:    parent,
		id:        id,
	}

	for i, childBuilder := range builder.children {
		child := childBuilder.build(route, strconv.Itoa(len(route.children)))
		route.children = append(route.children, child)
		if i == 0 {
			route.defaultChild = child
		}
	}

	return route
}
