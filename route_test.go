package gong

import (
	"testing"

	"github.com/troygilman/gong/assert"
)

func TestRouteBasic(t *testing.T) {
	comp := MockComponent{
		view: textTemplComponent{"view"},
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil)

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(""))
	assert.Equals(t, []Route{}, route.Children())
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
	comp := MockComponent{
		action: textTemplComponent{"action"},
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil)

	testRender(t, route, gongContext{action: true, id: "mock"}, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := MockComponent{
		action:     loaderTemplComponent{},
		loaderData: "action",
	}

	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil)

	testRender(t, route, gongContext{action: true, id: "mock"}, "action")
}
