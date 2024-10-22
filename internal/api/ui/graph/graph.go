package graph

import (
	"context"
	"github.com/labstack/echo/v4"
	"stagebase/internal/common"
)

const (
	ctxKey        = "graph"
	paramNameType = "graph_type"
	paramNameNode = "graph_node"
)

type Graph struct {
	graphType common.GraphType
	node      string
}

func Extract(e echo.Context) (any, any) {
	graphType := common.GraphType(e.QueryParam(paramNameType))
	if graphType == "" {
		graphType = common.GraphTypeAll
	}

	node := e.QueryParam(paramNameNode)

	return ctxKey, Graph{
		graphType: graphType,
		node:      node,
	}
}

func FromCtx(ctx context.Context) (Graph, bool) {
	g, ok := ctx.Value(ctxKey).(Graph)
	return g, ok
}
