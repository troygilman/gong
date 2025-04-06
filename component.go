package gong

import (
	"context"
	"io"
	"reflect"
	"strings"
)

type Component struct {
	kind     string
	view     View
	loader   Loader
	action   Action
	children map[string]Component
}

func NewComponent(kind string, view View) Component {
	component := Component{
		kind:     kind,
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

func (component Component) Find(kind string) (Component, bool) {
	return component.find(strings.Split(kind, "_"))
}

func (component Component) find(kind []string) (Component, bool) {
	if len(kind) == 0 || len(kind) == 1 && kind[0] == "" {
		return component, true
	}
	if child, ok := component.children[kind[0]]; ok {
		return child.find(kind[1:])
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

func (component Component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.loader = component.loader

	if gCtx.action {
		gCtx.action = false
		return render(ctx, gCtx, w, component.action.Action())
	}

	if gCtx.kind == "" {
		gCtx.kind = component.kind
	} else {
		gCtx.kind += "_" + component.kind
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
				children[child.kind] = child
			}
		}
	}
	return children
}
