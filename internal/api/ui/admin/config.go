package admin

import (
	"stagebase/internal/api/common"
	"stagebase/internal/core"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func Configure(g gong.Group) {
	gong.NewGroup(g, "/admin/:project_id",
		func(e *echo.Group) {
			e.GET("", gong.NewTemplHandler(AdminProducer))
			e.GET("/sse", handleConnectSseProjectAdmin)
		},
		gong.ContextExtractors(
			common.ExtractProjectId,
		),
		gong.ContextLogHandler(
			common.ParamProjectId,
		),
		common.LogRequestMiddleware(),
	)
}

func handleConnectSseProjectAdmin(e echo.Context) error {
	ctx := e.Request().Context()
	return common.HandleServerSentEvents(e, core.ListenerTypeMetrics, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case core.MetricGroup:
				if err := common.WriteServerSentEvent(
					ctx,
					e.Response(),
					sseTypeUpdateMetricLatency,
					metric(sse.Avg.Latency, sseTypeUpdateMetricLatency),
				); err != nil {
					return err
				}
				if err := common.WriteServerSentEvent(
					ctx,
					e.Response(),
					sseTypeUpdateMetricThroughput,
					metric(sse.Avg.Throughput, sseTypeUpdateMetricThroughput),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
