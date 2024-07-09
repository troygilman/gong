package dolly

import (
	"errors"
	"log"

	"github.com/anthdm/hollywood/actor"
)

func ForwardRequest(ctx *actor.Context, pid *actor.PID, msg any) {
	ctx.Engine().SendWithSender(pid, msg, ctx.Sender())
}

func RespondWithError(ctx *actor.Context, err error) error {
	ctx.Respond(err)
	return err
}

func HandleResponse[T any](response *actor.Response) (T, error) {
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
		log.Printf("ERROR: %s - %T - %s", ctx.PID().GetID(), ctx.Message(), err.Error())
	}
}

func LogActorMsg(ctx *actor.Context) {
	log.Printf("MESSAGE: %s - %T", ctx.PID().GetID(), ctx.Message())
}
