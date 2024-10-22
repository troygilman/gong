package persistance

import (
	"context"
	"stagebase/internal/model"
	"testing"

	"github.com/google/uuid"
)

func benchmarkCreateEvents(n int, numEvents int, db Database) error {
	projectID := uuid.MustParse("2db8b7bc-5237-4a01-bf7a-d0339da782d9")
	for range n {
		events := make([]model.Event, numEvents)
		for idx := range numEvents {
			events[idx] = model.Event{
				ProjectID: projectID.String(),
			}
		}
		if err := db.CreateEvents(context.Background(), events); err != nil {
			return err
		}
	}
	return nil
}

func BenchmarkCreateEvents_10(b *testing.B) {
	db, err := NewQuestDatabase(nil)
	if err != nil {
		b.Error(err)
		return
	}
	if err := benchmarkCreateEvents(b.N, 10, db); err != nil {
		b.Error(err)
		return
	}
}

func BenchmarkCreateEvents_100(b *testing.B) {
	db, err := NewQuestDatabase(nil)
	if err != nil {
		b.Error(err)
		return
	}
	if err := benchmarkCreateEvents(b.N, 100, db); err != nil {
		b.Error(err)
		return
	}
}

func BenchmarkCreateEvents_1000(b *testing.B) {
	db, err := NewQuestDatabase(nil)
	if err != nil {
		b.Error(err)
		return
	}
	if err := benchmarkCreateEvents(b.N, 1000, db); err != nil {
		b.Error(err)
		return
	}
}

func BenchmarkCreateEvents_10000(b *testing.B) {
	db, err := NewQuestDatabase(nil)
	if err != nil {
		b.Error(err)
		return
	}
	if err := benchmarkCreateEvents(b.N, 10000, db); err != nil {
		b.Error(err)
		return
	}
}
