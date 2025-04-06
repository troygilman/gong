package gong

type RouteBuilder struct {
	path      string
	component Component
	children  []RouteBuilder
}

func NewRoute(path string, view View) RouteBuilder {
	return RouteBuilder{
		path:      path,
		component: NewComponent("", view),
	}
}

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
