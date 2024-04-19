package handler

import (
	"fmt"
	"gong/domain"

	"github.com/labstack/echo/v4"
)

func mustApp(ctx echo.Context) domain.App {
	app, ok := ctx.Get("app").(domain.App)
	if !ok {
		panic(fmt.Sprintln("App does not exist in context"))
	}
	return app
}
