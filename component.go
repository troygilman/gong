package gong

import (
	"context"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

const (
	idDelimeter = "_"
)

type gongComponent struct {
	view     View
	loader   Loader
	action   Action
	head     Head
	id       string
	children map[string]Component
}

// New creates a new Component instance with the specified view.
// It automatically scans the view for any child components and sets up
// optional interfaces (Loader, Action, Head) if the view implements them.
func NewComponent(view View, opts ...ComponentOption) Component {
	component := gongComponent{
		id:       nextComponentID(),
		view:     view,
		children: make(map[string]Component),
	}

	for _, opt := range opts {
		component = opt(component)
	}

	component.scanViewForActions()

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
func (component gongComponent) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.Component = component

	if gCtx.ComponentID == "" {
		gCtx.ComponentID = component.id
	} else {
		gCtx.ComponentID += idDelimeter + component.id
	}

	return render(ctx, gCtx, w, component.view.View())
}

func (component gongComponent) View() templ.Component {
	return component
}

func (component gongComponent) Action() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if component.action == nil {
			return nil
		}
		gCtx := getContext(ctx)
		gCtx.Component = component
		return render(ctx, gCtx, w, component.action.Action())
	})
}

func (component gongComponent) Loader(ctx context.Context) any {
	if component.loader == nil {
		return nil
	}
	return component.loader.Loader(ctx)
}

func (component gongComponent) Head() templ.Component {
	if component.head == nil {
		return defaultHead()
	}
	return component.head.Head()
}

func (component gongComponent) ID() string {
	return component.id
}

// Find searches for a child component with the specified ID.
// The ID can be a simple identifier or a path of IDs separated by the delimiter.
// Returns the found component and a boolean indicating if it was found.
func (component gongComponent) Find(idStr string) (Component, bool) {
	id := strings.Split(idStr, idDelimeter)
	if len(id) > 0 && id[0] == component.id {
		if len(id) == 1 {
			return component, true
		}
		if child, ok := component.children[id[1]]; ok {
			return child.Find(strings.Join(id[1:], idDelimeter))
		}
	}
	return gongComponent{}, false
}

// WithLoaderFunc sets a loader function for the component.
// The loader function will be called to fetch data before rendering.
// Returns the modified component for method chaining.
func (component gongComponent) WithLoaderFunc(loader LoaderFunc) Component {
	component.loader = loader
	return component
}

// WithLoaderData sets static data for the component's loader.
// This is a convenience method for components that don't need dynamic data loading.
// Returns the modified component for method chaining.
func (component gongComponent) WithLoaderData(data any) Component {
	component.loader = LoaderFunc(func(ctx context.Context) any {
		return data
	})
	return component
}

// Option is a function type for configuring components with the options pattern.
// It takes a gongComponent and returns a modified one.
type ComponentOption func(gongComponent) gongComponent

// WithID sets a custom ID for the component.
// This ID is used for component identification and event handling.
// Returns the modified component for method chaining.
func withID(id string) ComponentOption {
	return func(gc gongComponent) gongComponent {
		gc.id = id
		return gc
	}
}

func (component gongComponent) scanViewForActions() {
	v := reflect.ValueOf(component.view)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			if child, ok := field.Interface().(Component); ok {
				component.children[child.ID()] = child
			}
		}
	}
}

var _nextComponentID = 0

func nextComponentID() string {
	id := strconv.Itoa(_nextComponentID)
	_nextComponentID++
	return id
}
