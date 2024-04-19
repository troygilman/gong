package domain

import "github.com/google/uuid"

type persistentActor interface {
	load(id uuid.UUID) error
	create(id uuid.UUID) error
}
