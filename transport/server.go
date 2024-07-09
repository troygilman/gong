package transport

import (
	"gong/core"
	"io/fs"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer(app core.App, static fs.FS) error {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	e.Any("/debug/pprof/", makeHandler(pprof.Index))
	e.Any("/debug/pprof/cmdline", makeHandler(pprof.Cmdline))
	e.Any("/debug/pprof/profile", makeHandler(pprof.Profile))
	e.Any("/debug/pprof/symbol", makeHandler(pprof.Symbol))
	e.Any("/debug/pprof/trace", makeHandler(pprof.Trace))

	e.Use(middleware.RequestID())
	e.Use(middlewareLogRequest())
	e.Use(middlewareWithApp(app))

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	bindProjectHandlers(e)

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

func makeHandler(handler http.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler(c.Response().Writer, c.Request())
		return nil
	}
}
