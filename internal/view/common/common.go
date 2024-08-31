package common

import (
	"context"
	"fmt"
	"gong/internal/common"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func FormatProjectUrl(ctx context.Context) string {
	url := fmt.Sprintf("/project/%s", ctx.Value(common.CtxKeyProjectID).(uuid.UUID).String())
	return string(templ.URL(url))
}

const (
	IdActorPID  = "actor-pid"
	IdTab       = "tab"
	IdFilter    = "events-filter"
	IdFlow      = "flow"
	IdFlowNode  = "flow-node"
	IdFlowRange = "flow-range"
)

var StandardIncludes = []string{
	IdActorPID,
	IdTab,
	IdFilter,
	IdFlow,
	IdFlowNode,
	IdFlowRange,
}
