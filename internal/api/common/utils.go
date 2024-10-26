package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"stagebase/internal/core"
	"stagebase/internal/model"
	"stagebase/pkg/gong"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func HandleServerSentEvents(e echo.Context, listenerType core.ListenerType, f func(sses []any) error) error {
	ctx := e.Request().Context()

	var (
		app       core.App
		projectId uuid.UUID
		requestId string
		filter    model.Filter
	)

	if err := gong.Bind(
		ctx,
		gong.NewBinding(ParamApp, &app),
		gong.NewBinding(ParamProjectId, &projectId),
		gong.NewBinding(ParamRequestId, &requestId),
		gong.NewBinding(ParamFilter, &filter),
	); err != nil {
		return err
	}

	flusher, ok := e.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	e.Response().Header().Set("Content-Type", "text/event-stream")

	sseChan, err := app.ListenStart(ctx, projectId, requestId, listenerType, filter)
	if err != nil {
		return err
	}

	defer func() {
		if err := app.ListenStop(ctx, requestId); err != nil {
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

func WriteServerSentEvent(ctx context.Context, w io.Writer, event string, data any) error {
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
