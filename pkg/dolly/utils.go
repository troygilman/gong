package dolly

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/anthdm/hollywood/actor"
)

func ForwardRequest(ctx *actor.Context, pid *actor.PID, msg any) {
	ctx.Engine().SendWithSender(pid, msg, ctx.Sender())
}

func RespondWithError(ctx *actor.Context, err error) error {
	ctx.Respond(err)
	return err
}

func Result[T any](response *actor.Response) (T, error) {
	var none T
	result, err := response.Result()
	if err != nil {
		return none, err
	}

	switch result := result.(type) {
	case T:
		return result, nil
	case error:
		return none, result
	default:
		return none, errors.New("invalid result")
	}
}

func LogActorError(ctx *actor.Context, err error) {
	if err != nil {
		slog.ErrorContext(ctx.Context(), err.Error(), "actor", ctx.PID().GetID(), "type", fmt.Sprintf("%T", ctx.Message()))
	}
}

func LogActorMsg(ctx *actor.Context) {
	slog.InfoContext(ctx.Context(), "MESSAGE", "actor", ctx.PID().GetID(), "type", fmt.Sprintf("%T", ctx.Message()))
}
