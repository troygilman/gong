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
	path         string
	component    gong.Component
	children     []gong.Route
	defaultChild gong.Route
	parent       gong.Route
	id           string
}

func (route *gongRoute) Render(ctx context.Context, w io.Writer) error {
	gCtx := gctx.GetContext(ctx)
	gCtx.Route = route
	// gCtx.Path = buildRealPath(route, gCtx.request)

	// log.Printf("Rendering Route: %+v\n", gCtx)
	if len(route.children) > 0 {
		depth := route.Depth()
		if len(gCtx.RouteID) > depth {
			index := int(gCtx.RouteID[depth] - '0')
			gCtx.ChildRoute = route.Child(index)
		} else {
			gCtx.ChildRoute = route.children[0]
		}
	}

	if gCtx.Link {
		parent := route.Parent()
		if parent == nil {
			panic("could not find parent")
		}
		gCtx.Route = parent
		gCtx.ChildRoute = route
		gCtx.Link = false
		return util.Render(ctx, gCtx, w, outlet.New().WithRoute(route).WithOOB(true))
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

func (route *gongRoute) Child(index int) gong.Route {
	if index < 0 || index >= len(route.children) {
		return nil
	}
	return route.children[index]
}

func (route *gongRoute) Find(id string) gong.Route {
	var r gong.Route = route
	for _, index := range id {
		r = r.Child(int(index - '0'))
	}
	return r
}

func (route *gongRoute) ID() string {
	if route.parent != nil {
		return route.parent.ID() + route.id
	}
	return route.id
}

func (route *gongRoute) NumChildren() int {
	return len(route.children)
}

func (route *gongRoute) Parent() gong.Route {
	return route.parent
}

func (route *gongRoute) Root() gong.Route {
	if route.parent == nil {
		return route
	}
	return route.parent
}

func (route *gongRoute) Component() gong.Component {
	return route.component
}

func (route *gongRoute) Path() string {
	return route.path
}

func (route *gongRoute) FullPath() string {
	if route.parent != nil {
		return route.parent.FullPath() + route.path
	}
	return route.path
}

func (route *gongRoute) Depth() int {
	if route.parent == nil {
		return 0
	} else {
		return route.parent.Depth() + 1
	}
}
