package gong

import (
	"context"
	"io"
	"reflect"
	"strconv"

	"github.com/a-h/templ"
)

const (
	idDelimeter = "_"
)

type Component interface {
	templ.Component
	View
	Action
	Loader
	Index
	ID() string
	Find(id []string) (Component, bool)
	WithLoaderFunc(loader LoaderFunc) Component
	WithLoaderData(data any) Component
	WithID(id string) Component
}

// Component represents a UI component in the Gong framework.
// It encapsulates a view, optional loader, action, and head elements,
// along with any child components that may be part of its structure.
type gongComponent struct {
	view     View
	loader   Loader
	action   Action
	index    Index
	id       string
	children map[string]Component
}

// NewComponent creates a new Component instance with the specified view.
// It automatically scans the view for any child components and sets up
// optional interfaces (Loader, Action, Head) if the view implements them.
func NewComponent(view View) Component {
	component := gongComponent{
		id:       nextComponentID(),
		view:     view,
		children: make(map[string]Component),
	}
	component.scanViewForActions()

	if loader, ok := view.(Loader); ok {
		component.loader = loader
	}

	if action, ok := view.(Action); ok {
		component.action = action
	}

	if index, ok := view.(Index); ok {
		component.index = index
	}

	return component
}

// Render writes the component's HTML representation to the provided writer.
// It handles both normal rendering and action execution based on the context.
// Returns an error if rendering fails.
func (component gongComponent) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.component = component

	if gCtx.action {
		gCtx.action = false
		return render(ctx, gCtx, w, component.Action())
	}

	if gCtx.componentID == "" {
		gCtx.componentID = component.id
	} else {
		gCtx.componentID += idDelimeter + component.id
	}

	return render(ctx, gCtx, w, component.View())
}

func (component gongComponent) View() templ.Component {
	return component.view.View()
}

func (component gongComponent) Action() templ.Component {
	if component.action == nil {
		return nil
	}
	return component.action.Action()
}

func (component gongComponent) Loader(ctx context.Context) any {
	if component.loader == nil {
		return nil
	}
	return component.loader.Loader(ctx)
}

func (component gongComponent) Head() templ.Component {
	if component.index == nil {
		return DefaultHead()
	}
	return component.index.Head()
}

func (component gongComponent) HtmlAttrs() templ.Attributes {
	if component.index == nil {
		return nil
	}
	return component.index.HtmlAttrs()
}

func (component gongComponent) ID() string {
	return component.id
}

// Find searches for a child component with the specified ID.
// The ID can be a simple identifier or a path of IDs separated by the delimiter.
// Returns the found component and a boolean indicating if it was found.
func (component gongComponent) Find(id []string) (Component, bool) {
	if len(id) > 0 && id[0] == component.id {
		if len(id) == 1 {
			return component, true
		}
		if child, ok := component.children[id[1]]; ok {
			return child.Find(id[1:])
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

// WithID sets a custom ID for the component.
// This ID is used for component identification and event handling.
// Returns the modified component for method chaining.
func (component gongComponent) WithID(id string) Component {
	component.id = id
	return component
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
