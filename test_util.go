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

type testComponent struct {
	view       templ.Component
	action     templ.Component
	loaderData any
}

func (c testComponent) View() templ.Component {
	if c.view != nil {
		return c.view
	} else {
		return testTemplComponent{"view"}
	}
}

func (c testComponent) Action() templ.Component {
	return c.action
}

func (c testComponent) Loader(ctx context.Context) any {
	return c.loaderData
}

type testTemplComponent struct {
	text string
}

func (c testTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, c.text)
	return err
}

type testLoaderTemplComponent struct{}

func (c testLoaderTemplComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, LoaderData[string](ctx))
	return err
}

type testParentComponent struct {
	Child Component
}

func (c testParentComponent) View() templ.Component {
	return nil
}

type testRoute struct {
	path string
}

func (r testRoute) Path() string {
	return r.path
}

func (r testRoute) Component() Component {
	return nil
}

func (r testRoute) Render(ctx context.Context, w io.Writer) error {
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
func newTestRequest(method string, url string) *http.Request {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return r
}
