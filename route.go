package gong

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type RouteOption func(Route) Route

func WithChildren(children ...Route) RouteOption {
	return func(r Route) Route {
		r.children = children
		return r
	}
}

type Route struct {
	path      string
	component Component
	children  []Route
}

func (r Route) Path() string {
	return r.path
}

func NewRoute(path string, component Component, opts ...RouteOption) Route {
	route := Route{
		path:      path,
		component: component,
	}
	for _, opt := range opts {
		route = opt(route)
	}
	return route
}

func (route Route) newNode(parent *routeNode, id string) *routeNode {
	node := &routeNode{
		route:  route,
		parent: parent,
		id:     id,
		depth:  len(id),
	}
	if parent != nil {
		node.path = parent.path + route.path
	}
	for _, child := range route.children {
		node.children = append(node.children, child.newNode(node, id+strconv.Itoa(len(node.children))))
	}
	return node
}

type routeNode struct {
	route    Route
	path     string
	id       string
	depth    int
	parent   *routeNode
	children []*routeNode
}

func (node *routeNode) Render(ctx context.Context, w io.Writer) error {
	gCtx := getContext(ctx)
	gCtx.Node = node
	gCtx.ChildRouteIndex = 0

	// log.Printf("Rendering Route: %+v\n", gCtx)
	if len(node.children) > 0 {
		if len(gCtx.RequestRouteID) > node.depth {
			gCtx.ChildRouteIndex = int(gCtx.RequestRouteID[node.depth] - '0')
		}
	}

	if gCtx.Link {
		gCtx.Link = false
		return render(ctx, gCtx, w, Outlet(withOOB(true)))
	}

	if gCtx.Action {
		component, ok := node.route.component.Find(gCtx.ComponentID)
		if !ok {
			panic(fmt.Sprintf("could not find component with id %s in route %s", gCtx.ComponentID, node.route.path))
		}
		return render(ctx, gCtx, w, component.Action())
	}

	gCtx.ComponentID = ""
	return render(ctx, gCtx, w, node.route.component.View())
}

func (node *routeNode) find(id string) *routeNode {
	var n *routeNode = node
	for _, index := range id {
		n = n.children[int(index-'0')]
	}
	return n
}
