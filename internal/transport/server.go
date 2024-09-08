package transport

import (
	"io/fs"
	"log"
	"net/http"
	"net/http/pprof"
	"stagebase/internal/common"
	"stagebase/pkg/gong"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer(app common.App, static fs.FS, port int) (*echo.Echo, <-chan error) {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	e.Any("/debug/pprof/", makeHandler(pprof.Index))
	e.Any("/debug/pprof/cmdline", makeHandler(pprof.Cmdline))
	e.Any("/debug/pprof/profile", makeHandler(pprof.Profile))
	e.Any("/debug/pprof/symbol", makeHandler(pprof.Symbol))
	e.Any("/debug/pprof/trace", makeHandler(pprof.Trace))

	e.Use(
		withHeaders(),
		middleware.RequestID(),
		withRequestID(),
		gong.ContextLogHandler(
			common.CtxKeyRequestID,
		),
		withApp(app),
	)

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	bindProjectHandlers(e)

	errChan := make(chan error)
	go func() {
		errChan <- e.Start(":" + strconv.Itoa(port))
	}()
	return e, errChan
}

func httpErrorHandler(err error, ctx echo.Context) {
	switch err := err.(type) {
	default:
		log.Printf("ERROR: %s %s - %s", ctx.Request().Method, ctx.Request().RequestURI, err.Error())
	}
}

func makeHandler(handler http.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler(c.Response().Writer, c.Request())
		return nil
	}
}
