package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"stagebase/internal/api/common"
	v1 "stagebase/internal/api/v1"
	"stagebase/internal/api/view"
	"stagebase/internal/core"
	"stagebase/internal/model"
	"stagebase/pkg/gong"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	datastar "github.com/starfederation/datastar/sdk/go"
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
		common.WithApp(app),
	)

	e.StaticFS("/static", echo.MustSubFS(static, "static"))

	v1.Configure(e)
	// project.Configure(e)
	// admin.Configure(e)

	e.GET("/project/:project_id", func(c echo.Context) error {
		ctx := c.Request().Context()
		projectID := ctx.Value(common.ParamProjectId).(uuid.UUID)
		return view.Home(projectID.String()).Render(ctx, c.Response())
	},
		gong.Extractors(
			gong.NewExtractor(common.ExtractProjectId, gong.WithLogging()),
		),
	)

	e.GET("/project/:project_id/events", func(c echo.Context) error {
		ctx := c.Request().Context()
		sse := datastar.NewSSE(c.Response(), c.Request())
		app := ctx.Value(common.ParamApp).(core.App)
		projectID := ctx.Value(common.ParamProjectId).(uuid.UUID)
		requestID := ctx.Value(common.ParamRequestId).(string)

		stream, err := app.ListenStart(ctx, projectID, requestID, core.ListenerTypeEvents, model.Filter{Level: slog.LevelDebug})
		if err != nil {
			return err
		}

		for events := range stream {
			for _, event := range events {
				switch event := event.(type) {
				case core.SseNewEventsMsg:
					sse.MergeFragmentTempl(view.Event(event.Actor))

				}
			}
		}
		return nil
	},
		gong.Extractors(
			gong.NewExtractor(common.ExtractProjectId, gong.WithLogging()),
		),
	)

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
