package gong

import (
	"context"
	"io"
)

type Route interface {
	Route(path string, view View, f func(r Route))
}

type route struct {
	gong         *Gong
	path         string
	view         View
	actions      map[string]Action
	children     map[string]*route
	defaultChild *route
	parent       *route
}

func (r *route) Route(path string, view View, f func(r Route)) {
	newRoute := &route{
		gong:     r.gong,
		view:     view,
		path:     r.path + path,
		actions:  make(map[string]Action),
		children: make(map[string]*route),
		parent:   r,
	}
	r.addChild(newRoute.path, newRoute)
	r.gong.handleRoute(newRoute)
	if f != nil {
		f(newRoute)
	}
}

func (r *route) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.route = r

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

func (r *route) addChild(path string, child *route) {
	r.children[path] = child
	if r.defaultChild == nil {
		r.defaultChild = child
	}
	if r.parent != nil {
		r.parent.addChild(path, r)
	}
}

func (r *route) getRoute(path string) *route {
	if r.path == path {
		return r
	}
	return r.children[path].getRoute(path)
}

func (r *route) getRoot() *route {
	if r.parent == nil {
		return r
	}
	return r.parent
}
