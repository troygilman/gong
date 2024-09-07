package transport

import (
	"stagebase/internal/common"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func extractProjectID(ctx echo.Context) (common.CtxKey, any) {
	projectID, err := uuid.Parse(ctx.Param(string(common.CtxKeyProjectID)))
	if err != nil {
		return "", nil
	}
	return common.CtxKeyProjectID, projectID
}

func extractActorPID(ctx echo.Context) (common.CtxKey, any) {
	actorPID := ctx.QueryParam(common.ParamPID)
	if actorPID == "" {
		return "", nil
	}
	return common.CtxKeyPID, actorPID
}

func extractTab(ctx echo.Context) (common.CtxKey, any) {
	tab := ctx.QueryParam(common.ParamTab)
	if tab == "" {
		tab = "Events"
	}
	return common.CtxKeyTab, tab
}

func extractFlowType(ctx echo.Context) (common.CtxKey, any) {
	flowType := ctx.QueryParam(common.ParamFlow)
	if flowType == "" {
		flowType = "Sources"
	}
	return common.CtxKeyFlow, flowType
}

func extractFlowNode(ctx echo.Context) (common.CtxKey, any) {
	flowNode := ctx.QueryParam(common.ParamFlowNode)
	if flowNode == "" {
		return "", nil
	}
	return common.CtxKeyFlowNode, flowNode
}

func extractFlowRange(ctx echo.Context) (common.CtxKey, any) {
	var flowRange common.Range
	if err := flowRange.Parse(ctx.QueryParam(common.ParamFlowRange)); err != nil {
		return "", nil
	}
	return common.CtxKeyFlowRange, flowRange
}

func extractFilter(ctx echo.Context) (common.CtxKey, any) {
	return common.CtxKeyFilter, common.NewFilter(ctx)
}

func extractTimeStart(ctx echo.Context) (common.CtxKey, any) {
	timeStartStr := ctx.QueryParam(common.ParamTimeStart)
	timeStartInt64, err := strconv.ParseInt(timeStartStr, 10, 64)
	if err != nil {
		return "", nil
	}
	return common.CtxKeyTimeStart, time.Unix(0, timeStartInt64)
}

func extractPageIndex(ctx echo.Context) (common.CtxKey, any) {
	pageIdxStr := ctx.QueryParam(common.ParamPageIndex)
	pageIdx, err := strconv.Atoi(pageIdxStr)
	if err != nil {
		return "", nil
	}
	return common.CtxKeyPageIndex, pageIdx
}
