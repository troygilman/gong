package model

import (
	"time"

	"github.com/google/uuid"
)

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
