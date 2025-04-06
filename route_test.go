package gong

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/troygilman/gong/assert"
)

type MockComponent struct {
	view       templ.Component
	action     templ.Component
	loaderData any
}

func (mc MockComponent) View() templ.Component {
	return mc.view
}

func (mc MockComponent) Action() templ.Component {
	return mc.action
}

func (mc MockComponent) Loader(ctx context.Context) any {
	return mc.loaderData
}

type textTemplComponent struct {
	text string
}

func (c textTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, c.text)
	return err
}

type loaderTemplComponent struct{}

func (c loaderTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, LoaderData[string](ctx))
	return err
}

func TestRouteBasic(t *testing.T) {
	comp := MockComponent{
		view: textTemplComponent{"view"},
	}

	route := NewRoute("/", comp).build(nil)

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(""))
	assert.Equals(t, []Route{}, route.Children())
	assert.Equals(t, route, route.Root())
	assert.Equals(t, Component{
		view:     comp,
		action:   comp,
		loader:   comp,
		children: make(map[string]Component),
	}, route.Component())

	testRouteRender(t, route, gongContext{}, "view")
}

func TestRouteAction(t *testing.T) {
	comp := MockComponent{
		action: textTemplComponent{"action"},
	}

	route := NewRoute("/", comp).build(nil)

	testRouteRender(t, route, gongContext{action: true}, "action")
}

func TestRouteAction_withLoader(t *testing.T) {
	comp := MockComponent{
		action:     loaderTemplComponent{},
		loaderData: "action",
	}

	route := NewRoute("/", comp).build(nil)

	testRouteRender(t, route, gongContext{action: true}, "action")
}

func testRouteRender(t *testing.T, route Route, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, route)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}
