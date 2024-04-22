package handler

import (
	"bytes"
	"fmt"
	"gong/domain"
	"log"
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

func renderView(ctx echo.Context, view templ.Component) error {
	return view.Render(ctx.Request().Context(), ctx.Response())
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
	log.Println(sb.String())
	if _, err := fmt.Fprint(ctx.Response().Writer, sb.String()); err != nil {
		return err
	}

	return nil
}

func requestID(ctx echo.Context) string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}
