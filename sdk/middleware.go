package sdk

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"reflect"
	"strconv"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func StagebaseMiddleware() actor.MiddlewareFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			event := Event{
				PID:       ctx.PID().String(),
				Timestamp: time.Now(),
			}

			if ctx.Sender() != nil {
				event.Sender = ctx.Sender().String()
			}

			rType := reflect.TypeOf(ctx.Message())
			if rType != nil {
				event.Type = rType.String()
			}

			switch msg := ctx.Message().(type) {
			case actor.EventLogger:
				level, logMsg, attr := msg.Log()

				if attrStr, err := encodeAttrMap(createAttrMap(attr)); err == nil {
					event.Attrs = attrStr
				}

				event.Level = strconv.Itoa(int(level))
				event.Msg = logMsg
			case actor.Initialized, actor.Started, actor.Stopped:
				event.Level = strconv.Itoa(int(slog.LevelInfo))
			}

			ctx.Send(stagebasePID, event)

			next(ctx)
		}
	}
}

func createAttrMap(attr []any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(attr); i += 2 {
		var val any = nil
		if i+1 < len(attr) {
			val = attr[i+1]
		}
		switch key := attr[i].(type) {
		case string:
			m[key] = val
		}
	}
	return m
}

func encodeAttrMap(attr map[string]any) (string, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(attr); err != nil {
		return "", err
	}
	return buffer.String(), nil

}
