package gong

import (
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestRouteBasic(t *testing.T) {
	comp := mockComponent{
		view: textTemplComponent{"view"},
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil, "")

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(0))
	assert.Equals(t, 0, route.NumChildren())
	assert.Equals(t, route, route.Root())
	assert.Equals(t, Component{
		id:       "mock",
		view:     comp,
		action:   comp,
		loader:   comp,
		children: make(map[string]Component),
	}, route.Component())

	testRender(t, route, gongContext{}, "view")
}

func TestRouteRenderAction(t *testing.T) {
	comp := mockComponent{
		action: textTemplComponent{"action"},
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil, "")

	testRender(t, route, gongContext{action: true, componentID: "mock"}, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := mockComponent{
		action:     loaderTemplComponent{},
		loaderData: "action",
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil, "")

	testRender(t, route, gongContext{action: true, componentID: "mock"}, "action")
}
