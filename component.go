package gong

import (
	"context"
	"io"
)

type Component struct {
	kind   string
	view   View
	loader Loader
}

func NewComponent(kind string, view View) Component {
	component := Component{
		kind: kind,
		view: view,
	}
	if loader, ok := view.(Loader); ok {
		component.loader = loader
	}
	return component
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
	gCtx.action = false
	gCtx.loader = component.loader
	if gCtx.kind == "" {
		gCtx.kind = component.kind
	} else {
		gCtx.kind += "_" + component.kind
	}
	ctx = context.WithValue(ctx, contextKey, gCtx)

	return component.view.View().Render(ctx, w)
}
