package gong

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/a-h/templ"
	"github.com/troygilman/gong/internal/assert"
)

type MockComponent struct {
	MockView       templ.Component
	MockAction     templ.Component
	MockLoaderData any
}

func (mc MockComponent) View() templ.Component {
	if mc.MockView != nil {
		return mc.MockView
	} else {
		return TextTemplComponent{"view"}
	}
}

func (mc MockComponent) Action() templ.Component {
	return mc.MockAction
}

func (mc MockComponent) Loader(ctx context.Context) any {
	return mc.MockLoaderData
}

type TextTemplComponent struct {
	Text string
}

func (c TextTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, c.Text)
	return err
}

type LoaderTemplComponent struct{}

func (c LoaderTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, LoaderData[string](ctx))
	return err
}

type ParentComponent struct {
	Child Component
}

func (pc ParentComponent) View() templ.Component {
	return nil
}

type MockRoute struct {
	MockPath string
}

func (mock MockRoute) Child(id int) Route {
	return nil
}

func (mock MockRoute) NumChildren() int {
	return 0
}

func (mock MockRoute) Find(id string) (Route, int) {
	return nil, 0
}

func (mock MockRoute) Path() string {
	return mock.MockPath
}

func (mock MockRoute) Component() Component {
	return nil
}

func (mock MockRoute) Render(ctx context.Context, w io.Writer) error {
	return nil
}

// testRender is a helper function for testing templ component rendering.
// It renders the given component with the provided context and asserts
// that the output matches the expected string.
func TestRender(t *testing.T, c templ.Component, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, c)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}

// newRequest creates a new HTTP request with the given method and URL.
// It panics if the request creation fails. This is a helper function
// for creating request objects in tests.
func NewRequest(method string, url string) *http.Request {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return r
}
