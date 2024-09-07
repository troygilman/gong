package view

import (
	"context"
	"fmt"
	"stagebase/internal/common"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func FormatProjectUrl(ctx context.Context) string {
	url := fmt.Sprintf("/project/%s", ctx.Value(common.CtxKeyProjectID).(uuid.UUID).String())
	return string(templ.URL(url))
}

const (
	IdentifierPID = "actor-pid"
	IdTab         = "tab"
	IdFlow        = "flow"
	IdFlowNode    = "flow-node"
	IdFlowRange   = "flow-range"
)

var StandardIncludes = []string{
	IdentifierPID,
	IdTab,
	IdFlow,
	IdFlowNode,
	IdFlowRange,
}
