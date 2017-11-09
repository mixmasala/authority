package authority

import (
	"context"
	"time"

	"github.com/katzenpost/authority"
	"github.com/katzenpost/core/epochtime"

	"github.com/jonboulle/clockwork"
)

const (
	tilExchange = 120 * time.Minute
	tilTabulate = (127 * time.Minute) + (30 * time.Second)
)

type EpochScheduler struct {
	ctx context.Context
	c   clockwork.Clock
	et  *epochtime.Clock
}

func NewEpochScheduler(ctx context.Context, c clockwork.Clock) *EpochScheduler {
	return &EpochScheduler{ctx, c, epochtime.New(c)}
}

func (e *EpochScheduler) Run(sm authority.StateMachine) {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			e.Runone(sm)
		}
	}
}

func (e *EpochScheduler) Runone(sm authority.StateMachine) {
	_, elapsed, til := e.et.Now()

	var tq []<-chan time.Time

	if elapsed < tilExchange {
		dt := tilExchange - elapsed
		tq = append(tq, e.c.After(dt))
	}

	if elapsed < tilTabulate {
		dt := tilTabulate - elapsed
		tq = append(tq, e.c.After(dt))
	}

	dt := til
	tq = append(tq, e.c.After(dt))

	for _, ch := range tq {
		select {
		case <-e.ctx.Done():
			return
		case <-ch:
			sm.Advance()
		}
	}
}
