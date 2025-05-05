package test_util

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"net/http"
// 	"testing"

// 	"github.com/a-h/templ"
// 	"github.com/troygilman/gong"
// 	"github.com/troygilman/gong/hooks"
// 	"github.com/troygilman/gong/internal/assert"
// 	"github.com/troygilman/gong/internal/gctx"
// 	"github.com/troygilman/gong/internal/util"
// )

// type MockComponent struct {
// 	View       templ.Component
// 	Action     templ.Component
// 	LoaderData any
// }

// func (mc MockComponent) View() templ.Component {
// 	if mc.view != nil {
// 		return mc.view
// 	} else {
// 		return TextTemplComponent{"view"}
// 	}
// }

// func (mc MockComponent) Action() templ.Component {
// 	return mc.action
// }

// func (mc MockComponent) Loader(ctx context.Context) any {
// 	return mc.loaderData
// }

// type TextTemplComponent struct {
// 	text string
// }

// func (c TextTemplComponent) Render(ctx context.Context, w io.Writer) error {
// 	_, err := io.WriteString(w, c.text)
// 	return err
// }

// type LoaderTemplComponent struct{}

// func (c LoaderTemplComponent) Render(ctx context.Context, w io.Writer) error {
// 	_, err := io.WriteString(w, hooks.LoaderData[string](ctx))
// 	return err
// }

// type ParentComponent struct {
// 	Child gong.Component
// }

// func (pc ParentComponent) View() templ.Component {
// 	return nil
// }

// type MockRoute struct {
// 	path string
// }

// func (mock MockRoute) Child(id int) gong.Route {
// 	return nil
// }

// func (mock MockRoute) NumChildren() int {
// 	return 0
// }

// func (mock MockRoute) Parent() gong.Route {
// 	return nil
// }

// func (mock MockRoute) Depth() int {
// 	return 0
// }

// func (mock MockRoute) Root() gong.Route {
// 	return mock
// }

// func (mock MockRoute) Find(id string) gong.Route {
// 	return nil
// }

// func (mock MockRoute) Path() string {
// 	return mock.path
// }

// func (mock MockRoute) FullPath() string {
// 	return ""
// }

// func (mock MockRoute) Component() gong.Component {
// 	return nil
// }

// func (mock MockRoute) Render(ctx context.Context, w io.Writer) error {
// 	return nil
// }

// // testRender is a helper function for testing templ component rendering.
// // It renders the given component with the provided context and asserts
// // that the output matches the expected string.
// func TestRender(t *testing.T, c templ.Component, gCtx gctx.Context, expected string) {
// 	buffer := bytes.NewBuffer([]byte{})
// 	err := util.Render(context.Background(), gCtx, buffer, c)
// 	assert.Equals(t, nil, err)
// 	assert.Equals(t, expected, buffer.String())
// }

// // newRequest creates a new HTTP request with the given method and URL.
// // It panics if the request creation fails. This is a helper function
// // for creating request objects in tests.
// func NewRequest(method string, url string) *http.Request {
// 	r, err := http.NewRequest(method, url, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return r
// }
