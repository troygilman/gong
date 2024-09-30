package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"stagebase/internal/common"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func mustApp(e echo.Context) common.App {
	app, ok := e.Get("app").(common.App)
	if !ok {
		panic(fmt.Sprintln("App does not exist in context"))
	}
	return app
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

func requestID(e echo.Context) string {
	return e.Response().Header().Get(echo.HeaderXRequestID)
}

func mustProjectID(e echo.Context) uuid.UUID {
	return e.Request().Context().Value(common.CtxKeyProjectID).(uuid.UUID)
}

func actorPID(e echo.Context) (string, bool) {
	actorPID, ok := e.Request().Context().Value(common.CtxKeyPID).(string)
	return actorPID, ok
}

func sse(e echo.Context, listenerType common.ListenerType, f func(sses []any) error) error {
	ctx := e.Request().Context()
	app := mustApp(e)

	flusher, ok := e.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	e.Response().Header().Set("Content-Type", "text/event-stream")

	requestID := requestID(e)

	projectID := mustProjectID(e)

	filter, ok := e.Request().Context().Value(common.CtxKeyFilter).(common.Filter)

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
