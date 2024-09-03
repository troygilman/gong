package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"stagebase/internal"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func mustApp(ctx echo.Context) internal.App {
	app, ok := ctx.Get("app").(internal.App)
	if !ok {
		panic(fmt.Sprintln("App does not exist in context"))
	}
	return app
}

func renderView(echoCtx echo.Context, ctx context.Context, view templ.Component) error {
	return view.Render(ctx, echoCtx.Response())
}

func writeServerSentEvent(ctx context.Context, w io.Writer, event string, data any) error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("event: %s\n", event))

	var d string
	switch data := data.(type) {
	case templ.Component:
		buff := bytes.NewBuffer([]byte{})
		if err := data.Render(ctx, buff); err != nil {
			return err
		}
		d = strings.ReplaceAll(buff.String(), "\n", "")
	default:
		if data == nil {
			d = ""
		}
	}

	sb.WriteString(fmt.Sprintf("data: %s\n\n", d))
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return err
	}

	return nil
}

func requestID(ctx echo.Context) string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}

func mustProjectID(ctx echo.Context) uuid.UUID {
	return ctx.Request().Context().Value(internal.CtxKeyProjectID).(uuid.UUID)
}

func actorPID(ctx echo.Context) (string, bool) {
	actorPID, ok := ctx.Request().Context().Value(internal.CtxKeyPID).(string)
	return actorPID, ok
}

func tab(ctx echo.Context) (string, bool) {
	tab, ok := ctx.Request().Context().Value(internal.CtxKeyTab).(string)
	return tab, ok
}

func filter(ctx echo.Context) (internal.Filter, bool) {
	filter, ok := ctx.Request().Context().Value(internal.CtxKeyFilter).(internal.Filter)
	return filter, ok
}

func sse(eCtx echo.Context, listenerType internal.ListenerType, f func(sses []any) error) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	flusher, ok := eCtx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	eCtx.Response().Header().Set("Content-Type", "text/event-stream")

	requestID := requestID(eCtx)

	projectID := mustProjectID(eCtx)

	filter, _ := filter(eCtx)

	sseChan, err := app.ListenStart(ctx, projectID, requestID, listenerType, filter)
	if err != nil {
		return err
	}

	defer func() {
		if err := app.ListenStop(ctx, requestID); err != nil {
			log.Println(err)
		}
		// flush the channel until closed
		for range sseChan {
		}
	}()

	for {
		select {
		case sses := <-sseChan:
			if err := f(sses); err != nil {
				return err
			}
			flusher.Flush()
		}
	}
}
