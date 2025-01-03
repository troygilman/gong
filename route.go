package gong

import (
	"context"
	"io"
)

type Route struct {
	gong    *Gong
	path    string
	view    View
	actions map[string]Action
}

func (r Route) Route(path string, view View, f func(r Route)) {
	r.path += path
	r.view = view
	r.gong.handleRoute(r)
	f(Route{
		gong: r.gong,
		path: r.path,
	})
}

func (r Route) Path() string {
	return r.path
}

func (r Route) View() View {
	return r.view
}

func (route Route) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)

	if gCtx.action {
		if action, ok := route.actions[gCtx.kind]; ok {
			gCtx.loader = nil
			if loader, ok := action.(Loader); ok {
				gCtx.loader = loader
			}
			return render(ctx, gCtx, w, action.Action())
		}
		if action, ok := route.View().(Action); ok {
			return render(ctx, gCtx, w, action.Action())
		}
		return nil
	}

	return render(ctx, gCtx, w, route.View().View())
}
