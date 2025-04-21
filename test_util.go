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

type mockComponent struct {
	view       templ.Component
	action     templ.Component
	loaderData any
}

func (mc mockComponent) View() templ.Component {
	if mc.view != nil {
		return mc.view
	} else {
		return textTemplComponent{"view"}
	}
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

func (mock mockRoute) Child(id int) Route {
	return nil
}

func (mock mockRoute) NumChildren() int {
	return 0
}

func (mock mockRoute) Parent() Route {
	return nil
}

func (mock mockRoute) Depth() int {
	return 0
}

func (mock mockRoute) ID() string {
	return ""
}

func (mock mockRoute) Root() Route {
	return mock
}

func (mock mockRoute) Path() string {
	return mock.path
}

func (mock mockRoute) FullPath() string {
	return ""
}

func (mock mockRoute) Component() Component {
	return nil
}

func (mock mockRoute) Render(ctx context.Context, w io.Writer) error {
	return nil
}

// testRender is a helper function for testing templ component rendering.
// It renders the given component with the provided context and asserts
// that the output matches the expected string.
func testRender(t *testing.T, c templ.Component, gCtx gongContext, expected string) {
	buffer := bytes.NewBuffer([]byte{})
	err := render(context.Background(), gCtx, buffer, c)
	assert.Equals(t, nil, err)
	assert.Equals(t, expected, buffer.String())
}

// newRequest creates a new HTTP request with the given method and URL.
// It panics if the request creation fails. This is a helper function
// for creating request objects in tests.
func newRequest(method string, url string) *http.Request {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return r
}
