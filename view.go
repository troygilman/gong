package gong

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

type viewComponent struct {
	kind   string
	view   View
	action bool
	config componentConfig
}

func ViewComponent(kind string, view View, opts ...ComponentOption) templ.Component {
	c := viewComponent{
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

func (c viewComponent) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.action = c.action
	gCtx.loader = c.config.loader
	gCtx.kind = c.kind
	ctx = context.WithValue(ctx, contextKey, gCtx)

	return c.view.View().Render(ctx, w)
}
