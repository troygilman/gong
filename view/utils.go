package view

import (
	"context"
	"fmt"
	"gong/types"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func formatProjectUrl(ctx context.Context, path string) string {
	url := fmt.Sprintf("/project/%s%s", ctx.Value(types.CtxProject).(uuid.UUID).String(), path)
	return string(templ.URL(url))
}

const includes = "#actor-pid, #tab, #events-filter"
