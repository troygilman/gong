package transport

import (
	"stagebase/internal/common"
	"stagebase/pkg/gong"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func extractProjectID(e echo.Context) (gong.Key, any) {
	projectID, err := uuid.Parse(e.Param(string(common.CtxKeyProjectID)))
	if err != nil {
		return "", nil
	}
	return common.CtxKeyProjectID, projectID
}

func extractActorPID(e echo.Context) (gong.Key, any) {
	actorPID := e.QueryParam(common.ParamPID)
	if actorPID == "" {
		return "", nil
	}
	return common.CtxKeyPID, actorPID
}

func extractTab(e echo.Context) (gong.Key, any) {
	tab := e.QueryParam(common.ParamTab)
	if tab == "" {
		tab = "Events"
	}
	return common.CtxKeyTab, tab
}

func extractFlowType(e echo.Context) (gong.Key, any) {
	flowType := e.QueryParam(common.ParamFlow)
	if flowType == "" {
		flowType = "Sources"
	}
	return common.CtxKeyFlow, flowType
}

func extractFlowNode(e echo.Context) (gong.Key, any) {
	flowNode := e.QueryParam(common.ParamFlowNode)
	if flowNode == "" {
		return "", nil
	}
	return common.CtxKeyFlowNode, flowNode
}

func extractFlowRange(e echo.Context) (gong.Key, any) {
	var flowRange common.Range
	if err := flowRange.Parse(e.QueryParam(common.ParamFlowRange)); err != nil {
		return "", nil
	}
	return common.CtxKeyFlowRange, flowRange
}

func extractFilter(e echo.Context) (gong.Key, any) {
	return common.CtxKeyFilter, common.NewFilter(e)
}

func extractTimeStart(e echo.Context) (gong.Key, any) {
	timeStartStr := e.QueryParam(common.ParamTimeStart)
	timeStartInt64, err := strconv.ParseInt(timeStartStr, 10, 64)
	if err != nil {
		return "", nil
	}
	return common.CtxKeyTimeStart, time.Unix(0, timeStartInt64)
}

func extractPageIndex(e echo.Context) (gong.Key, any) {
	pageIdxStr := e.QueryParam(common.ParamPageIndex)
	pageIdx, err := strconv.Atoi(pageIdxStr)
	if err != nil {
		return "", nil
	}
	return common.CtxKeyPageIndex, pageIdx
}
