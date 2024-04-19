package handler

import (
	"fmt"
	"gong/domain"

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
