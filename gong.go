package gong

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
)

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

// Component represents a complete UI component in the Gong framework.
// It combines several interfaces (View, Action, Loader, Head) to provide
// a unified component model with rendering, action handling, data loading,
// and head element capabilities.
type Component interface {
	templ.Component
	View
	Action
	Loader
	Head
	// ID returns the unique identifier for this component.
	ID() string
	// Find searches for a child component with the given ID.
	// Returns the found component and a boolean indicating success.
	Find(id string) (Component, bool)
	// WithLoaderFunc attaches a loader function to the component for data loading.
	// Returns the modified component for method chaining.
	WithLoaderFunc(loader LoaderFunc) Component
	// WithLoaderData sets static data for the component.
	// Returns the modified component for method chaining.
	WithLoaderData(data any) Component
}

// Route represents a route in the application's routing tree.
// It defines the interface for handling component routing and rendering.
type Route interface {
	templ.Component

	// Child returns the child route for the given path.
	// If no exact match is found and a default child exists, returns the default child.
	// Returns nil if no matching route is found.
	Child(int) Route

	Find(string) (Route, int)

	// NumChildren returns the number of direct child routes of this route.
	NumChildren() int

	// Path returns the path segment that this route represents.
	Path() string

	// Component returns the component associated with this route.
	Component() Component
}

// TriggerAfterSwap creates an HTMX event trigger that fires after a swap operation
// completes for an element with the specified ID.
func TriggerAfterSwap(id string) string {
	return fmt.Sprintf("htmx:afterSwap[detail.target.id === '%s'] from:body", id)
}

// TriggerAfterSwapOOB creates an HTMX event trigger that fires after an out-of-band
// swap operation completes for an element with the specified ID.
func TriggerAfterSwapOOB(id string) string {
	return fmt.Sprintf("htmx:oobAfterSwap[detail.target.id === '%s'] from:body", id)
}

// Error creates a templ.Component that returns the provided error when rendered.
// This is useful for propagating errors through the component tree.
func Error(err error) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return err
	})
}

// ErrorHandler is a function type for handling errors that occur during component rendering.
// It provides a way to implement custom error handling strategies.
type ErrorHandler func(context.Context, error)
