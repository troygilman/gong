package main

import (
	"embed"
	"gong/domain"
	"gong/example"
	"gong/handler"
	"log"
	"os"

	"net/http"
	"net/http/pprof"

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

	e.Any("/debug/pprof/", makeHandler(pprof.Index))
	e.Any("/debug/pprof/cmdline", makeHandler(pprof.Cmdline))
	e.Any("/debug/pprof/profile", makeHandler(pprof.Profile))
	e.Any("/debug/pprof/symbol", makeHandler(pprof.Symbol))
	e.Any("/debug/pprof/trace", makeHandler(pprof.Trace))

	e.Use(middleware.RequestID())
	e.Use(handler.MiddlewareLogRequest())
	e.Use(handler.MiddlewareWithApp(app))

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	newGroup(e, "/project/:projectID", func(e *echo.Group) {
		e.GET("", handler.GetProject)
		e.GET("/events", handler.GetProjectEvents)
		e.GET("/sse", handler.ConnectProjectSSE)
		e.GET("/chart", handler.GetActorChart)
		e.GET("/tooltip", handler.GetActorChartTooltip)
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

func makeHandler(handler http.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler(c.Response().Writer, c.Request())
		return nil
	}
}
