package view

import (
	"context"
	"fmt"
	"gong/types"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func formatProjectUrl(ctx context.Context) string {
	url := fmt.Sprintf("/project/%s", ctx.Value(types.CtxProject).(uuid.UUID).String())
	return string(templ.URL(url))
}

const includes = "#actor-pid, #tab, #events-filter"
