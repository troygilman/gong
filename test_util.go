package gong

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/troygilman/gong/internal/assert"
)

type mockComponent struct {
	view       templ.Component
	action     templ.Component
	loaderData any
}

func (mc mockComponent) View() templ.Component {
	return mc.view
}

func (mc mockComponent) Action() templ.Component {
	return mc.action
}

func (mc mockComponent) Loader(ctx context.Context) any {
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

type parentComponent struct {
	Child Component
}

func (pc parentComponent) View() templ.Component {
	return nil
}

type mockRoute struct {
	path string
}

func (mock mockRoute) Child(path string) Route {
	return nil
}

func (mock mockRoute) Children() []Route {
	return []Route{}
}

func (mock mockRoute) Root() Route {
	return mock
}

func (mock mockRoute) Path() string {
	return mock.path
}

func (mock mockRoute) Component() Component {
	return Component{}
}

func (mock mockRoute) Render(ctx context.Context, w io.Writer) error {
	return nil
}

func testRender(t *testing.T, c templ.Component, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, c)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}
