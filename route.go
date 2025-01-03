package gong

import (
	"context"
	"io"
)

type Route interface {
	Route(path string, view View, f func(r Route))
}

type route struct {
	gong    *Gong
	path    string
	view    View
	actions map[string]Action
}

func (r *route) Route(path string, view View, f func(r Route)) {
	newRoute := &route{
		gong:    r.gong,
		view:    Index{IndexView: view},
		path:    r.path + path,
		actions: make(map[string]Action),
	}
	r.gong.handleRoute(newRoute)
	f(newRoute)
}

func (r *route) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)

	if gCtx.action {
		if action, ok := r.actions[gCtx.kind]; ok {
			gCtx.loader = nil
			if loader, ok := action.(Loader); ok {
				gCtx.loader = loader
			}
			return render(ctx, gCtx, w, action.Action())
		}
		if action, ok := r.view.(Action); ok {
			return render(ctx, gCtx, w, action.Action())
		}
		return nil
	}

	return render(ctx, gCtx, w, r.view.View())
}
