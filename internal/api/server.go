package api

import (
	"io/fs"
	"net/http"
	"net/http/pprof"
	"stagebase/internal/api/common"
	"stagebase/internal/api/ui/project"
	v1 "stagebase/internal/api/v1"
	"stagebase/internal/core"
	"stagebase/pkg/gong"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer(app core.App, static fs.FS, port int) (*echo.Echo, <-chan error) {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	e.Any("/debug/pprof/", makeHandler(pprof.Index))
	e.Any("/debug/pprof/cmdline", makeHandler(pprof.Cmdline))
	e.Any("/debug/pprof/profile", makeHandler(pprof.Profile))
	e.Any("/debug/pprof/symbol", makeHandler(pprof.Symbol))
	e.Any("/debug/pprof/trace", makeHandler(pprof.Trace))

	e.Use(
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			//LogErrorFunc: newPanicError,
		}),
		common.WithHeadersMiddleware(),
		middleware.RequestID(),
		common.WithRequestIdMiddleware(),
		gong.ContextLogHandler(
			common.ParamRequestId,
		),
		common.WithApp(app),
	)

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	v1.Configure(e)
	project.Configure(e)

	errChan := make(chan error)
	go func() {
		errChan <- e.Start(":" + strconv.Itoa(port))
	}()
	return e, errChan
}

func httpErrorHandler(err error, e echo.Context) {
	switch err := err.(type) {
	default:
		ctx := e.Request().Context()
		logger := gong.Logger(ctx)
		logger.ErrorContext(ctx, err.Error(), "method", e.Request().Method, "uri", e.Request().RequestURI)
	}
}

func newPanicError(e echo.Context, err error, stack []byte) error {
	return err
}

func makeHandler(handler http.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler(c.Response().Writer, c.Request())
		return nil
	}
}
