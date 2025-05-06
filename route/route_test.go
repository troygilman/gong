package route

import (
	"net/http"
	"testing"

	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/internal/assert"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/test_util"
)

func TestRouteBasic(t *testing.T) {
	comp := test_util.MockComponent{
		MockView: test_util.TextTemplComponent{Text: "view"},
	}

	route := New("/", component.New(comp, component.WithID("mock")))

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(0))
	assert.Equals(t, 0, route.NumChildren())

	ctx := gctx.Context{
		Request: test_util.NewRequest(http.MethodGet, "/"),
	}

	test_util.TestRender(t, route, ctx, "view")
}

func TestRouteRenderAction(t *testing.T) {
	comp := test_util.MockComponent{
		MockAction: test_util.TextTemplComponent{Text: "action"},
	}

	route := New("/", component.New(comp, component.WithID("mock")))

	ctx := gctx.Context{
		Action:      true,
		ComponentID: "mock",
		Request:     test_util.NewRequest(http.MethodGet, "/"),
	}

	test_util.TestRender(t, route, ctx, "action")
}

func TestRouteRenderAction_withLoader(t *testing.T) {
	comp := test_util.MockComponent{
		MockAction:     test_util.LoaderTemplComponent{},
		MockLoaderData: "action",
	}

	route := New("/", component.New(comp, component.WithID("mock")))

	ctx := gctx.Context{
		Action:      true,
		ComponentID: "mock",
		Request:     test_util.NewRequest(http.MethodGet, "/"),
	}

	test_util.TestRender(t, route, ctx, "action")
}
