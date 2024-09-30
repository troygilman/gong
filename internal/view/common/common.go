package view

import (
	"context"
	"fmt"
	"stagebase/internal/common"
	"strings"

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
	IdFromTime    = "filter-from-time"
)

var StandardIncludes = []string{
	IdentifierPID,
	IdTab,
	IdFlow,
	IdFlowNode,
	IdFlowRange,
	IdFromTime,
}

func BuildIncludes(includes ...[]string) string {
	var result []string
	for _, include := range includes {
		result = append(result, include...)
	}
	for i, include := range result {
		result[i] = "#" + include
	}
	return strings.Join(result, ",")
}
