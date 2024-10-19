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

func extractPage(e echo.Context) (any, any) {
	page, err := strconv.Atoi(e.QueryParam(paramPage))
	if err != nil {
		return paramPage, 0
	}
	return paramPage, page
}

func mustGetPage(ctx context.Context) int {
	return common.MustGetHelper[int](ctx, paramPage)
}

func extractView(e echo.Context) (any, any) {
	view := e.Request().Header.Get("View")
	return paramView, view
}
