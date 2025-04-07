package gong

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

func (builder RouteBuilder) build(parent Route) Route {
	path := builder.path
	if parent != nil {
		path = parent.Path() + builder.path
	}

	route := &gongRoute{
		component: builder.component,
		path:      path,
		children:  make(map[string]Route),
		parent:    parent,
	}

	for i, childBuilder := range builder.children {
		childRoute := childBuilder.build(route)
		route.children[childRoute.Path()] = childRoute
		if i == 0 {
			route.defaultChild = childRoute
		}
	}

	return route
}
