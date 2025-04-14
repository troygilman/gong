package gong

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/troygilman/gong/internal/response_writer"
)

type contextKeyType int

const contextKey = contextKeyType(0)

// HTTP header keys used by Gong for request handling and routing
const (
	HeaderGongRequestType = "Gong-Request-Type"
	HeaderGongComponentID = "Gong-Component-ID"
	HeaderGongRoutePath   = "Gong-Route-Path"
)

// Request type constants used by Gong
const (
	GongRequestTypeAction = "action"
	GongRequestTypeRoute  = "route"
)

// HTMX trigger constants for component updates
const (
	// TriggerNone indicates no automatic trigger
	TriggerNone = "none"
	// TriggerLoad indicates the component should update on page load
	TriggerLoad  = "load"
	TriggerClick = "click"
)

// HTMX swap constants for content updates
const (
	// SwapNone indicates no content swapping
	SwapNone = "none"
	// SwapOuterHTML replaces the entire target element
	SwapOuterHTML = "outerHTML"
	// SwapInnerHTML replaces the content inside the target element
	SwapInnerHTML = "innerHTML"
	// SwapBeforeEnd appends content before the end of the target element
	SwapBeforeEnd = "beforeend"
)

// Gong is the main framework instance that handles routing and request processing.
// It implements the http.Handler interface and manages the application's routes.
type Gong struct {
	mux Mux
}

// New creates a new Gong instance with the specified HTTP mux.
// The mux is used for routing HTTP requests to the appropriate handlers.
func New(mux Mux) *Gong {
	return &Gong{
		mux: mux,
	}
}

// Routes registers one or more route builders with the Gong instance.
// Each route builder is built and set up with appropriate handlers.
// Returns the Gong instance for method chaining.
func (g *Gong) Routes(builders ...RouteBuilder) *Gong {
	for _, builder := range builders {
		g.setupRoute(builder.build(nil))
	}
	return g
}

func (g *Gong) setupRoute(route Route) {
	log.Printf("Route=%s\n", route.Path())
	g.mux.Handle(route.Path(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := response_writer.NewResponseWriter(w)
		requestType := r.Header.Get(HeaderGongRequestType)

		gCtx := gongContext{
			requestType: requestType,
			route:       route,
			path:        r.Header.Get(HeaderGongRoutePath),
			url:         r.URL.EscapedPath(),
			request:     r,
			writer:      writer,
			action:      requestType == GongRequestTypeAction,
			id:          r.Header.Get(HeaderGongComponentID),
		}

		var templComponent templ.Component
		switch requestType {
		case GongRequestTypeAction:
			if route.Path() != gCtx.path {
				gCtx.route = route.Child(gCtx.path)
			}
			templComponent = gCtx.route
		case GongRequestTypeRoute:
			gCtx.link = true
			templComponent = gCtx.route
		default:
			gCtx.path = route.Path()
			gCtx.route = route.Root()
			templComponent = index(gCtx.route)
		}

		if gCtx.route == nil {
			panic("route is nil")
		}

		if err := render(r.Context(), gCtx, writer, templComponent); err != nil {
			panic(err)
		}
		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}))

	for _, child := range route.Children() {
		g.setupRoute(child)
	}
}

// ServeHTTP implements the http.Handler interface.
// It delegates request handling to the underlying mux.
func (g *Gong) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

type gongContext struct {
	requestType string
	route       Route
	request     *http.Request
	writer      *response_writer.ResponseWriter
	path        string
	url         string
	action      bool
	link        bool
	loader      Loader
	id          string
}

// Mux is an interface for HTTP request multiplexing.
// It defines the methods required for routing HTTP requests.
type Mux interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Handle(path string, handler http.Handler)
}

// View is an interface for components that can render themselves.
// It defines the method for getting a templ component.
type View interface {
	View() templ.Component
}

// Loader is an interface for components that can load data.
// It defines the method for loading data in a context.
type Loader interface {
	Loader(ctx context.Context) any
}

// Action is an interface for components that can handle actions.
// It defines the method for getting an action component.
type Action interface {
	Action() templ.Component
}

// Head is an interface for components that can provide head elements.
// It defines the method for getting head elements.
type Head interface {
	Head() templ.Component
}

// LoaderFunc is a function type that implements the Loader interface.
// It allows for easy creation of loader functions.
type LoaderFunc func(ctx context.Context) any

// Loader implements the Loader interface for LoaderFunc.
func (f LoaderFunc) Loader(ctx context.Context) any {
	return f(ctx)
}

// RenderFunc is a function type that implements the templ.Component interface.
// It allows for custom rendering logic in components.
type RenderFunc func(ctx context.Context, w io.Writer) error

// Render implements the templ.Component interface for RenderFunc.
func (r RenderFunc) Render(ctx context.Context, w io.Writer) error {
	return r(ctx, w)
}
