package route

// import (
// 	"net/http"
// 	"testing"

// 	"github.com/troygilman/gong/internal/assert"
// )

// func TestRouteBasic(t *testing.T) {
// 	comp := mockComponent{
// 		view: textTemplComponent{"view"},
// 	}

// 	route := New("/", NewComponent(comp).WithID("mock")).build(nil, "")

// 	assert.Equals(t, "/", route.Path())
// 	assert.Equals(t, nil, route.Child(0))
// 	assert.Equals(t, 0, route.NumChildren())
// 	assert.Equals(t, route, route.Root())

// 	ctx := gongContext{
// 		request: newRequest(http.MethodGet, "/"),
// 	}

// 	testRender(t, route, ctx, "view")
// }

// func TestRouteRenderAction(t *testing.T) {
// 	comp := mockComponent{
// 		action: textTemplComponent{"action"},
// 	}

// 	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil, "")

// 	ctx := gongContext{
// 		action:      true,
// 		componentID: "mock",
// 		request:     newRequest(http.MethodGet, "/"),
// 	}

// 	testRender(t, route, ctx, "action")
// }

// func TestRouteRenderAction_withLoader(t *testing.T) {
// 	comp := mockComponent{
// 		action:     loaderTemplComponent{},
// 		loaderData: "action",
// 	}

// 	route := NewRoute("/", NewComponent(comp).WithID("mock")).build(nil, "")

// 	ctx := gongContext{
// 		action:      true,
// 		componentID: "mock",
// 		request:     newRequest(http.MethodGet, "/"),
// 	}

// 	testRender(t, route, ctx, "action")
// }
