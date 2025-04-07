package gong

import (
	"context"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const (
	idDelimeter = "_"
)

var nextID = 0

type Component struct {
	id       string
	view     View
	loader   Loader
	action   Action
	children map[string]Component
}

func NewComponent(view View) Component {
	id := strconv.Itoa(nextID)
	nextID++

	component := Component{
		id:       id,
		view:     view,
		children: scanViewForActions(view),
	}

	if loader, ok := view.(Loader); ok {
		component.loader = loader
	}

	if action, ok := view.(Action); ok {
		component.action = action
	}

	return component
}

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

func (component Component) WithLoaderFunc(loader LoaderFunc) Component {
	component.loader = loader
	return component
}

func (component Component) WithLoaderData(data any) Component {
	component.loader = LoaderFunc(func(ctx context.Context) any {
		return data
	})
	return component
}

func (component Component) WithID(id string) Component {
	component.id = id
	return component
}

func (component Component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.loader = component.loader

	if gCtx.action {
		gCtx.action = false
		return render(ctx, gCtx, w, component.action.Action())
	}

	if gCtx.id == "" {
		gCtx.id = component.id
	} else {
		gCtx.id += idDelimeter + component.id
	}

	return render(ctx, gCtx, w, component.view.View())
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
