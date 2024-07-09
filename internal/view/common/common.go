package common

import (
	"context"
	"fmt"
	"gong/common"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func FormatProjectUrl(ctx context.Context) string {
	url := fmt.Sprintf("/project/%s", ctx.Value(common.CtxProject).(uuid.UUID).String())
	return string(templ.URL(url))
}

const Includes = "#actor-pid, #tab, #events-filter, #flow, #flow-node, #flow-range"
