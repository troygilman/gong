package gong

import (
	"net/http"
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestRouteBasic(t *testing.T) {
	comp := MockComponent{
		MockView: TextTemplComponent{Text: "view"},
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(0))
	assert.Equals(t, 0, route.NumChildren())

	ctx := gongContext{
		Request: NewRequest(http.MethodGet, "/"),
	}

	TestRender(t, route, ctx, "view")
}

func TestRouteRenderAction(t *testing.T) {
	comp := MockComponent{
		MockAction: TextTemplComponent{Text: "action"},
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     NewRequest(http.MethodGet, "/"),
	}

	TestRender(t, route, ctx, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := MockComponent{
		MockAction:     LoaderTemplComponent{},
		MockLoaderData: "action",
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     NewRequest(http.MethodGet, "/"),
	}

	TestRender(t, route, ctx, "action")
}
