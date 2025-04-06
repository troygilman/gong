package gong

import "reflect"

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
		actions:   make(map[string]Action),
		children:  make(map[string]Route),
		parent:    parent,
	}

	scanViewForActions(route.actions, route.component.view, "")

	for i, childBuilder := range builder.children {
		childRoute := childBuilder.build(route)
		route.children[childRoute.Path()] = childRoute
		if i == 0 {
			route.defaultChild = childRoute
		}
	}

	return route
}

func scanViewForActions(actions map[string]Action, view View, kindPrefix string) {
	v := reflect.ValueOf(view)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			if component, ok := field.Interface().(Component); ok {
				kind := kindPrefix + component.kind
				actions[kind] = component.action
				scanViewForActions(actions, component.view, kind+"_")
			}
		}
	}
}
