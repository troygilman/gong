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

type ParentComponent struct {
	Child Component
}

func (pc ParentComponent) View() templ.Component {
	return nil
}

func testRender(t *testing.T, c templ.Component, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, c)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}
