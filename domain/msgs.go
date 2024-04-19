package domain

import "gong/sdk"

type MsgEvent struct {
	Event sdk.Event
}

type MsgGetProjectProps struct{}

type RspGetProjectProps struct {
	Props ProjectProps
}
