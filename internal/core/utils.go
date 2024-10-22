package core

import (
	"fmt"

	"github.com/anthdm/hollywood/actor"
)

func Request[Result any](app App, pid *actor.PID, msg any) (result Result, err error) {
	r, err := app.Request(pid, msg)
	if err != nil {
		return result, err
	}
	result, ok := r.(Result)
	if !ok {
		return result, fmt.Errorf("result type %T is not expected type %T", r, result)
	}
	return result, nil
}
