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
		uri:   "/",
		route: route,
		id:    "id",
	}

	testRender(t, component, ctx, `<form hx-post="/" hx-swap="none" hx-headers="{&#34;Gong-Request&#34;: &#34;action&#34;, &#34;Gong-Path&#34;: &#34;/&#34;, &#34;Gong-ID&#34;: &#34;id&#34;}" class="--templ-css-class-unknown-type"></form>`)
}
