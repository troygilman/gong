package form

import (
	"testing"

	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/test_util"
)

func TestForm(t *testing.T) {
	component := New()

	route := test_util.MockRoute{
		MockPath: "/",
	}

	ctx := gctx.Context{
		Path:        "/",
		Route:       route,
		ComponentID: "id",
	}

	test_util.TestRender(t, component, ctx, `<form hx-post hx-swap="none" hx-trigger="" hx-headers="{&#34;Gong-Request-Type&#34;: &#34;action&#34;, &#34;Gong-Route-ID&#34;: &#34;&#34;, &#34;Gong-Component-ID&#34;: &#34;id&#34;}"></form>`)
}
