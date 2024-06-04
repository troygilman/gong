package api

import (
	"bytes"
	"context"
	"fmt"
	"gong/common"
	"gong/core"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func mustApp(ctx echo.Context) core.App {
	app, ok := ctx.Get("app").(core.App)
	if !ok {
		panic(fmt.Sprintln("App does not exist in context"))
	}
	return app
}

func renderView(echoCtx echo.Context, ctx context.Context, view templ.Component) error {
	return view.Render(ctx, echoCtx.Response())
}

func writeServerSentEvent(ctx context.Context, w io.Writer, event string, data any) error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("event: %s\n", event))

	var d string
	switch data := data.(type) {
	case templ.Component:
		buff := bytes.NewBuffer([]byte{})
		if err := data.Render(ctx, buff); err != nil {
			return err
		}
		d = strings.ReplaceAll(buff.String(), "\n", "")
	default:
		if data == nil {
			d = ""
		}
	}

	sb.WriteString(fmt.Sprintf("data: %s\n\n", d))
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return err
	}

	return nil
}

func requestID(ctx echo.Context) string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}

func mustProjectID(ctx echo.Context) uuid.UUID {
	return ctx.Request().Context().Value(common.CtxProject).(uuid.UUID)
}

func actorPID(ctx echo.Context) (string, bool) {
	actorPID, ok := ctx.Request().Context().Value(common.CtxPID).(string)
	return actorPID, ok
}

func tab(ctx echo.Context) (string, bool) {
	tab, ok := ctx.Request().Context().Value(common.CtxTab).(string)
	return tab, ok
}
