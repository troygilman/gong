package project

import (
	"github.com/labstack/echo/v4"
)

const (
	paramPid      = "pid"
	paramFlow     = "flow"
	paramFlowNode = "flow_node"
	paramTab      = "tab"
	paramFilter   = "filter"
	paramPage     = "page"
	paramView     = "view"
)

func extractActorPID(e echo.Context) (any, any) {
	actorPID := e.QueryParam(paramPid)
	if actorPID == "" {
		return nil, nil
	}
	return paramPid, actorPID
}

func extractTab(e echo.Context) (any, any) {
	tab := e.QueryParam(paramTab)
	if tab == "" {
		tab = tabLabelOverview
	}
	return paramTab, tab
}

func extractFlowType(e echo.Context) (any, any) {
	flowType := e.QueryParam(paramFlow)
	if flowType == "" {
		flowType = "Sources"
	}
	return paramFlow, flowType
}

func extractFlowNode(e echo.Context) (any, any) {
	flowNode := e.QueryParam(paramFlowNode)
	if flowNode == "" {
		return nil, nil
	}
	return paramFlowNode, flowNode
}

func extractView(e echo.Context) (any, any) {
	view := e.Request().Header.Get("View")
	return paramView, view
}
