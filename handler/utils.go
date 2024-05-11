package handler

import (
	"bytes"
	"context"
	"fmt"
	"gong/domain"
	"gong/types"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func mustApp(ctx echo.Context) domain.App {
	app, ok := ctx.Get("app").(domain.App)
	if !ok {
		panic(fmt.Sprintln("App does not exist in context"))
	}
	return app
}

func renderView(echoCtx echo.Context, ctx context.Context, view templ.Component) error {
	return view.Render(ctx, echoCtx.Response())
}

func writeServerSentEvent(ctx echo.Context, event string, data any) error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("event: %s\n", event))

	var d string
	switch data := data.(type) {
	case templ.Component:
		buff := bytes.NewBuffer([]byte{})
		if err := data.Render(ctx.Request().Context(), buff); err != nil {
			return err
		}
		d = strings.ReplaceAll(buff.String(), "\n", "")
	default:
		if data == nil {
			d = ""
		}
	}

	sb.WriteString(fmt.Sprintf("data: %s\n\n", d))
	if _, err := fmt.Fprint(ctx.Response().Writer, sb.String()); err != nil {
		return err
	}

	return nil
}

func requestID(ctx echo.Context) string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}

type contextBuilder struct {
	ctx context.Context
}

func buildContext(ctx echo.Context) contextBuilder {
	return contextBuilder{
		ctx: ctx.Request().Context(),
	}
}

func (cb contextBuilder) with(key types.CtxKey, val any) contextBuilder {
	cb.ctx = context.WithValue(cb.ctx, key, val)
	return cb
}

func (cb contextBuilder) build() context.Context {
	return cb.ctx
}
