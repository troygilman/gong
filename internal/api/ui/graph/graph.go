package graph

import (
	"github.com/labstack/echo/v4"
)

const (
	ctxKey        = "graph"
	paramNameType = "graph_type"
	paramNameNode = "graph_node"
)

type GraphType string

const (
	GraphTypeAll     = "All"
	GraphTypeSources = "Sources"
	GraphTypeSinks   = "Sinks"
)

type Graph struct {
	graphType GraphType
	Node      string          `json:"node"`
	Nodes     []GraphNodeData `json:"nodes"`
	Links     []GraphLinkData `json:"links"`
}

type GraphNodeData struct {
	Id string `json:"id"`
}

type GraphLinkData struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func Extract(e echo.Context) (any, any) {
	graphType := GraphType(e.QueryParam(paramNameType))
	if graphType == "" {
		graphType = GraphTypeAll
	}

	node := e.QueryParam(paramNameNode)

	return ctxKey, Graph{
		graphType: graphType,
		Node:      node,
	}
}
