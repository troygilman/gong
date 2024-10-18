package project

import (
	"context"
	"stagebase/internal/api/common"
	"strconv"

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

func mustGetActorPid(ctx context.Context) string {
	return common.MustGetHelper[string](ctx, paramPid)
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

func extractPageIndex(e echo.Context) (any, any) {
	pageIdxStr := e.QueryParam(paramPage)
	pageIdx, err := strconv.Atoi(pageIdxStr)
	if err != nil {
		return nil, nil
	}
	return paramPage, pageIdx
}

func extractView(e echo.Context) (any, any) {
	view := e.Request().Header.Get("View")
	return paramView, view
}
