package sdk

import (
	"fmt"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func GongMiddleware() actor.MiddlewareFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			switch msg := ctx.Message().(type) {
			case actor.EventLogger:
				level, logMsg, attr := msg.Log()
				ctx.Send(gongPID, Event{
					MsgType:   fmt.Sprintf("%T", msg),
					Timestamp: time.Now(),
					Receiver:  ctx.PID().String(),
					Level:     level,
					Msg:       logMsg,
					Attr:      createAttrMap(attr),
				})
			}
			next(ctx)
		}
	}
}
