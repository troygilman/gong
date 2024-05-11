package main

import (
	"embed"
	"gong/domain"
	"gong/example"
	"gong/handler"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed static
var static embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		example.StartClient()
	} else {
		log.Fatal(run())
	}
}

func run() error {
	app, err := domain.NewDefaultApp()
	if err != nil {
		return err
	}

	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	e.Use(middleware.RequestID())
	e.Use(handler.MiddlewareLogRequest())
	e.Use(handler.MiddlewareWithApp(app))

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	newGroup(e, "/project/:projectID", func(e *echo.Group) {
		e.GET("", handler.GetProject)
		e.GET("/actor", handler.GetProjectActor)
		e.GET("/events", handler.GetProjectEvents)
		e.GET("/sse", handler.ConnectProjectSSE)
		e.GET("/connect", handler.ConnectProject)
	}, handler.MiddlewareProjectContext())

	return e.Start(":8080")
}

func httpErrorHandler(err error, ctx echo.Context) {
	switch err := err.(type) {
	default:
		log.Printf("ERROR: %s %s - %s", ctx.Request().Method, ctx.Request().RequestURI, err.Error())
	}
}

type groupable interface {
	Group(string, ...echo.MiddlewareFunc) *echo.Group
}

func newGroup(e groupable, path string, f func(e *echo.Group), middlewares ...echo.MiddlewareFunc) {
	group := e.Group(path, middlewares...)
	f(group)
}
