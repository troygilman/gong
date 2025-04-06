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
	view   templ.Component
	action templ.Component
}

func (mc MockComponent) View() templ.Component {
	return mc.view
}

func (mc MockComponent) Action() templ.Component {
	return mc.action
}

type textTemplComponent struct {
	text string
}

func (c textTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, c.text)
	return err
}

func TestRouteBasic(t *testing.T) {
	comp := MockComponent{}
	comp.view = textTemplComponent{"Hello World"}

	route := NewRoute("/", comp).build(nil)

	assert.Equals(t, "/", route.Path())
	assert.Equals(t, nil, route.Child(""))
	assert.Equals(t, []Route{}, route.Children())
	assert.Equals(t, route, route.Root())
	assert.Equals(t, Component{
		view:   comp,
		action: comp,
	}, route.Component())

	testRouteRender(t, route, gongContext{}, "Hello World")
}

func testRouteRender(t *testing.T, route Route, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, route)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}
