package gong

import (
	"net/http"
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestRouteBasic(t *testing.T) {
	comp := testComponent{
		view: testTemplComponent{text: "view"},
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(0))
	assert.Equals(t, 0, route.NumChildren())

	ctx := gongContext{
		Request: newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, route, ctx, "view")
}

func TestRouteRenderAction(t *testing.T) {
	comp := testComponent{
		action: testTemplComponent{text: "action"},
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, route, ctx, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := testComponent{
		action:     testLoaderTemplComponent{},
		loaderData: "action",
	}

	route := NewRoute("/", NewComponent(comp, ComponentWithID("mock")))

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, route, ctx, "action")
}
