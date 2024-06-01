package domain

import (
	"context"
	"gong/types"
	"testing"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func benchmarkCreateEvents(n int, createEvents func(context.Context, []types.Event) error) error {
	projectID := uuid.MustParse("2db8b7bc-5237-4a01-bf7a-d0339da782d9")

	events := make([]types.Event, n)
	for idx := range n {
		events[idx] = types.Event{
			ID:        ulid.Make(),
			ProjectID: projectID,
		}
	}

	return createEvents(context.Background(), events)
}

func BenchmarkCreateEvents_10(b *testing.B) {
	db := NewPostgresDatabase()
	for range b.N {
		if err := benchmarkCreateEvents(10, db.CreateEvents); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCreateEvents_100(b *testing.B) {
	db := NewPostgresDatabase()
	for range b.N {
		if err := benchmarkCreateEvents(100, db.CreateEvents); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCreateEvents_1000(b *testing.B) {
	db := NewPostgresDatabase()
	for range b.N {
		if err := benchmarkCreateEvents(1000, db.CreateEvents); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCreateEvents_10000(b *testing.B) {
	db := NewPostgresDatabase()
	for range b.N {
		if err := benchmarkCreateEvents(10000, db.CreateEvents); err != nil {
			b.Error(err)
		}
	}
}
