package model

import (
	"log/slog"
	"stagebase/sdk"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ProjectID string
	sdk.Event
}

func (e Event) SlogLevel() slog.Level {
	i, err := strconv.Atoi(e.Level)
	if err != nil {
		panic(err)
	}
	return slog.Level(i)
}

type ProjectProps struct {
	ID     uuid.UUID
	Actors map[string]Actor
}

type Actor struct {
	PID           string
	ProjectID     uuid.UUID
	NumEvents     int
	Alive         bool
	LastEventTime time.Time
}

type Filter struct {
	PID      *string
	Type     *string
	Sender   *string
	Receiver *string
	Level    slog.Level
}

func (filter Filter) ValidEvent(event Event) bool {
	if filter.PID != nil && *filter.PID != event.PID {
		return false
	}

	if filter.Type != nil && *filter.Type != event.Type {
		return false
	}

	if filter.Sender != nil && *filter.Sender != event.Sender {
		return false
	}

	if filter.Level > event.SlogLevel() {
		return false
	}

	return true
}

type SourcesSinksNodeInfo struct {
	PID      string
	MsgCount int64
}

type GraphNode struct {
	Id string
}

type GraphLink struct {
	Source string
	Target string
}
