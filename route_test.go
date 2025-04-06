package gong

import (
	"net/http"
	"testing"

	"github.com/a-h/templ"
)

type MockComponent struct {
}

func (mc MockComponent) View() templ.Component {
	return nil
}

func (mc MockComponent) Action() templ.Component {
	return nil
}

type MockMux struct {
}

func (mm MockMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (mm MockMux) Handle(path string, handler http.Handler) {

}

func TestRoute(t *testing.T) {
	// mux := MockMux{}

	// route := NewRoute("/", MockComponent{}).build(mux, nil)

	// route.Render(ctx context.Context, w io.Writer)
}
