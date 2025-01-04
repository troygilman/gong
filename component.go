package gong

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

type component struct {
	kind   string
	view   View
	action bool
	config componentConfig
}

func Component(kind string, view View, opts ...ComponentOption) templ.Component {
	c := component{
		kind: kind,
		view: view,
	}
	if loader, ok := view.(Loader); ok {
		c.config.loader = loader
	}
	for _, opt := range opts {
		c.config = opt(c.config)
	}
	return c
}

func (c component) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = c.action
	gCtx.loader = c.config.loader
	if gCtx.kind == "" {
		gCtx.kind = c.kind
	} else {
		gCtx.kind += "_" + c.kind
	}
	ctx = context.WithValue(ctx, contextKey, gCtx)

	return c.view.View().Render(ctx, w)
}

type componentConfig struct {
	loader Loader
}

type ComponentOption func(c componentConfig) componentConfig

func ComponentWithLoaderData(data any) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.loader = LoaderFunc(func(ctx context.Context) any {
			return data
		})
		return c
	}
}
