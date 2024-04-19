package main

import (
	"gong/domain"
	"gong/example"
	"gong/handler"
	"log"
	"os"

	"github.com/labstack/echo/v4"
)

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

	e.Use(handler.MiddlewareLogRequest())

	e.Use(handler.MiddlewareWithApp(app))

	e.GET("project/:projectID", handler.GetProject)
	e.GET("project/:projectID/connect", handler.ConnectProject)

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

func newGroup(e groupable, path string, f func(group *echo.Group), middlewares ...echo.MiddlewareFunc) {
	group := e.Group(path, middlewares...)
	f(group)
}
