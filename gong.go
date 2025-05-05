package gong

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

type contextKeyType int

const contextKey = contextKeyType(0)

// HTTP header keys used by Gong for request handling and routing
const (
	HeaderGongRequestType = "Gong-Request-Type"
	HeaderGongComponentID = "Gong-Component-ID"
	HeaderGongRouteID     = "Gong-Route-ID"
)

// Request type constants used by Gong
const (
	GongRequestTypeAction = "action"
	GongRequestTypeLink   = "link"
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
	SwapInnerHTML   = "innerHTML"
	SwapBeforeBegin = "beforebegin"
	// SwapBeforeEnd appends content before the end of the target element
	SwapBeforeEnd = "beforeend"
)

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

type Component interface {
	templ.Component
	View
	Action
	Loader
	Head
	ID() string
	Find(id string) (Component, bool)
	WithLoaderFunc(loader LoaderFunc) Component
	WithLoaderData(data any) Component
	WithID(id string) Component
}

// Route represents a route in the application's routing tree.
// It defines the interface for handling component routing and rendering.
type Route interface {
	templ.Component

	// Child returns the child route for the given path.
	// If no exact match is found and a default child exists, returns the default child.
	// Returns nil if no matching route is found.
	Child(int) Route

	Find(string) Route

	// NumChildren returns the number of direct child routes of this route.
	NumChildren() int

	// Parent returns the parent route of this route.
	Parent() Route

	// Root returns the root route of the routing tree.
	Root() Route

	// Path returns the path segment that this route represents.
	Path() string

	// FullPath returns the full path of this route, including all parent paths.
	FullPath() string

	// ID returns the unique identifier for this route.
	ID() string

	// Depth returns the depth of this route in the routing tree.
	Depth() int

	// Component returns the component associated with this route.
	Component() Component
}
