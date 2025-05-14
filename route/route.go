package route

import (
	"context"
	"fmt"
	"io"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/util"
	"github.com/troygilman/gong/outlet"
)

type gongRoute struct {
	path      string
	component gong.Component
	children  []gong.Route
}

func New(path string, component gong.Component, opts ...Option) gong.Route {
	route := gongRoute{
		path:      path,
		component: component,
	}

	for _, opt := range opts {
		route = opt(route)
	}

	return route
}

func (route gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := gctx.GetContext(ctx)
	gCtx.Route = route
	gCtx.ChildRouteIndex = 0

	// log.Printf("Rendering Route: %+v\n", gCtx)
	if len(route.children) > 0 {
		if len(gCtx.RequestRouteID) > gCtx.Depth {
			gCtx.ChildRouteIndex = int(gCtx.RequestRouteID[gCtx.Depth] - '0')
		}
	}

	if gCtx.Link {
		gCtx.Link = false
		return util.Render(ctx, gCtx, w, outlet.New(outlet.WithOOB(true)))
	}

	if gCtx.Action {
		component, ok := route.component.Find(gCtx.ComponentID)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s in route %s", gCtx.ComponentID, route.path))
		}
		return util.Render(ctx, gCtx, w, component.Action())
	}

	gCtx.ComponentID = ""
	return util.Render(ctx, gCtx, w, route.component.View())
}

func (route gongRoute) Child(index int) gong.Route {
	if index < 0 || index >= len(route.children) {
		return nil
	}
	return route.children[index]
}

func (route gongRoute) Find(id string) (gong.Route, int) {
	var r gong.Route = route
	depth := 0
	for _, index := range id {
		r = r.Child(int(index - '0'))
		depth++
	}
	return r, depth
}

func (route gongRoute) NumChildren() int {
	return len(route.children)
}

func (route gongRoute) Component() gong.Component {
	return route.component
}

func (route gongRoute) Path() string {
	return route.path
}

type Option func(gongRoute) gongRoute

func WithChildren(children ...gong.Route) Option {
	return func(gr gongRoute) gongRoute {
		gr.children = children
		return gr
	}
}
