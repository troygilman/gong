package gong

import (
	"testing"
)

func TestForm(t *testing.T) {
	component := NewForm()
	route := mockRoute{
		path: "/",
	}
	ctx := gongContext{
		path:        "/",
		route:       route,
		componentID: "id",
	}

	testRender(t, component, ctx, `<form hx-post="/" hx-swap="none" hx-headers="{&#34;Gong-Request-Type&#34;: &#34;action&#34;, &#34;Gong-Route-ID&#34;: &#34;&#34;, &#34;Gong-Component-ID&#34;: &#34;id&#34;}"></form>`)
}
