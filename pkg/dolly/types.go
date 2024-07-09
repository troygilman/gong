package dolly

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

type PersistentActor interface {
	Load(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, id uuid.UUID) error
}

type PoisonTimer struct {
	c  chan struct{}
	mu sync.Mutex
}

func (pt *PoisonTimer) StartWithCallback(e *actor.Engine, pid *actor.PID, d time.Duration, c func() error) {
	pt.c = make(chan struct{})
	timer := time.NewTimer(d)
	go func() {
		for {
			select {
			case <-pt.c:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(d)
			case <-timer.C:
				locked := pt.mu.TryLock()
				if !locked {
					timer.Reset(d)
					continue
				}
				log.Printf("POISON: %s", pid.GetID())
				if c != nil {
					err := c()
					if err != nil {
						log.Printf("POISON ERROR: %s - %s", pid.GetID(), err.Error())
					}
				}
				e.Poison(pid)
				pt.c = nil
				pt.mu.Unlock()
				return
			}
		}
	}()
}

func (pt *PoisonTimer) Start(e *actor.Engine, pid *actor.PID, d time.Duration) {
	pt.StartWithCallback(e, pid, d, nil)
}

func (pt *PoisonTimer) Reset() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.c != nil {
		pt.c <- struct{}{}
	}
}
