package sdk

import (
	"fmt"
	"log/slog"

	"github.com/anthdm/hollywood/actor"
)

type Event struct {
	MsgType  string
	Receiver *actor.PID
	Level    slog.Level
	Msg      string
	Attr     []any
}

func MiddlewareGong() actor.MiddlewareFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			if gongActorPID != nil {
				switch msg := ctx.Message().(type) {
				case actor.EventLogger:
					level, logMsg, attr := msg.Log()
					ctx.Send(gongActorPID, Event{
						MsgType:  fmt.Sprintf("%T", msg),
						Receiver: ctx.PID(),
						Level:    level,
						Msg:      logMsg,
						Attr:     attr,
					})
				}
			}
			next(ctx)
		}
	}
}
