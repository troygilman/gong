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

	node := NewRoute("/", NewComponent(comp, withID("mock"))).newNode(nil, "")

	assert.Equals(t, "/", node.route.Path())

	ctx := gongContext{
		Request: newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, node, ctx, "view")
}

func TestRouteRenderAction(t *testing.T) {
	comp := testComponent{
		action: testTemplComponent{text: "action"},
	}

	node := NewRoute("/", NewComponent(comp, withID("mock"))).newNode(nil, "")

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, node, ctx, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := testComponent{
		action:     testLoaderTemplComponent{},
		loaderData: "action",
	}

	node := NewRoute("/", NewComponent(comp, withID("mock"))).newNode(nil, "")

	ctx := gongContext{
		Action:      true,
		ComponentID: "mock",
		Request:     newTestRequest(http.MethodGet, "/"),
	}

	testRender(t, node, ctx, "action")
}
