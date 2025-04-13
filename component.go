package gong

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const (
	idDelimeter = "_"
)

// Component represents a UI component in the Gong framework.
// It encapsulates a view, optional loader, action, and head elements,
// along with any child components that may be part of its structure.
type Component struct {
	id       string
	view     View
	loader   Loader
	action   Action
	head     Head
	children map[string]Component
}

// NewComponent creates a new Component instance with the specified view.
// It automatically scans the view for any child components and sets up
// optional interfaces (Loader, Action, Head) if the view implements them.
func NewComponent(view View) Component {
	component := Component{
		id:       nextComponentID(),
		view:     view,
		children: scanViewForActions(view),
	}

	if loader, ok := view.(Loader); ok {
		component.loader = loader
	}

	if action, ok := view.(Action); ok {
		component.action = action
	}

	if head, ok := view.(Head); ok {
		component.head = head
	}

	return component
}

// Render writes the component's HTML representation to the provided writer.
// It handles both normal rendering and action execution based on the context.
// Returns an error if rendering fails.
func (component Component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.loader = component.loader

	if gCtx.action {
		gCtx.action = false
		if component.action == nil {
			return fmt.Errorf("no action available")
		}
		return render(ctx, gCtx, w, component.action.Action())
	}

	if gCtx.id == "" {
		gCtx.id = component.id
	} else {
		gCtx.id += idDelimeter + component.id
	}

	return render(ctx, gCtx, w, component.view.View())
}

// Find searches for a child component with the specified ID.
// The ID can be a simple identifier or a path of IDs separated by the delimiter.
// Returns the found component and a boolean indicating if it was found.
func (component Component) Find(id string) (Component, bool) {
	return component.find(strings.Split(id, idDelimeter))
}

func (component Component) find(id []string) (Component, bool) {
	if len(id) > 0 && id[0] == component.id {
		if len(id) == 1 {
			return component, true
		}
		if child, ok := component.children[id[1]]; ok {
			return child.find(id[1:])
		}
	}
	return Component{}, false
}

// WithLoaderFunc sets a loader function for the component.
// The loader function will be called to fetch data before rendering.
// Returns the modified component for method chaining.
func (component Component) WithLoaderFunc(loader LoaderFunc) Component {
	component.loader = loader
	return component
}

// WithLoaderData sets static data for the component's loader.
// This is a convenience method for components that don't need dynamic data loading.
// Returns the modified component for method chaining.
func (component Component) WithLoaderData(data any) Component {
	component.loader = LoaderFunc(func(ctx context.Context) any {
		return data
	})
	return component
}

// WithID sets a custom ID for the component.
// This ID is used for component identification and event handling.
// Returns the modified component for method chaining.
func (component Component) WithID(id string) Component {
	component.id = id
	return component
}

func scanViewForActions(view View) map[string]Component {
	children := make(map[string]Component)
	v := reflect.ValueOf(view)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			if child, ok := field.Interface().(Component); ok {
				children[child.id] = child
			}
		}
	}
	return children
}

var _nextComponentID = 0

func nextComponentID() string {
	id := strconv.Itoa(_nextComponentID)
	_nextComponentID++
	return id
}
