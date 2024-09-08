package transport

import (
	"stagebase/internal/common"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func extractProjectID(e echo.Context) (any, any) {
	projectID, err := uuid.Parse(e.Param(string(common.CtxKeyProjectID)))
	if err != nil {
		return nil, nil
	}
	return common.CtxKeyProjectID, projectID
}

func extractActorPID(e echo.Context) (any, any) {
	actorPID := e.QueryParam(common.ParamPID)
	if actorPID == "" {
		return nil, nil
	}
	return common.CtxKeyPID, actorPID
}

func extractTab(e echo.Context) (any, any) {
	tab := e.QueryParam(common.ParamTab)
	if tab == "" {
		tab = "Events"
	}
	return common.CtxKeyTab, tab
}

func extractFlowType(e echo.Context) (any, any) {
	flowType := e.QueryParam(common.ParamFlow)
	if flowType == "" {
		flowType = "Sources"
	}
	return common.CtxKeyFlow, flowType
}

func extractFlowNode(e echo.Context) (any, any) {
	flowNode := e.QueryParam(common.ParamFlowNode)
	if flowNode == "" {
		return nil, nil
	}
	return common.CtxKeyFlowNode, flowNode
}

func extractFlowRange(e echo.Context) (any, any) {
	var flowRange common.Range
	if err := flowRange.Parse(e.QueryParam(common.ParamFlowRange)); err != nil {
		return nil, nil
	}
	return common.CtxKeyFlowRange, flowRange
}

func extractFilter(e echo.Context) (any, any) {
	return common.CtxKeyFilter, common.NewFilter(e)
}

func extractTimeStart(e echo.Context) (any, any) {
	timeStartStr := e.QueryParam(common.ParamTimeStart)
	timeStartInt64, err := strconv.ParseInt(timeStartStr, 10, 64)
	if err != nil {
		return nil, nil
	}
	return common.CtxKeyTimeStart, time.Unix(0, timeStartInt64)
}

func extractPageIndex(e echo.Context) (any, any) {
	pageIdxStr := e.QueryParam(common.ParamPageIndex)
	pageIdx, err := strconv.Atoi(pageIdxStr)
	if err != nil {
		return nil, nil
	}
	return common.CtxKeyPageIndex, pageIdx
}
