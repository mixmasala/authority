package authority

import (
	"time"
	"context"
	"fmt"

	"github.com/katzenpost/core/epochtime"

	"github.com/jonboulle/clockwork"
)

const (
	tilExchange = 120 * time.Minute
	tilTabulate = 150 * time.Minute
)

type Scheduler interface {
	Run(StateMachine)
}

type EpochScheduler struct {
	ctx context.Context
	c  clockwork.Clock
	et *epochtime.Clock
}

func NewEpochScheduler(ctx context.Context, c clockwork.Clock) *EpochScheduler {
	return &EpochScheduler{ctx, c, epochtime.New(c)}
}

func (e *EpochScheduler) Run(sm StateMachine) {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			e.Runone(sm)
		}
	}
}

func (e *EpochScheduler) Runone(sm StateMachine) {
	_, elapsed, til := e.et.Now()

	tq := make(chan (<-chan time.Time), 3)

	if elapsed < tilExchange {
		dt := tilExchange - elapsed
		fmt.Printf("waiting %v for exchange\n", dt)
		tq <- e.c.After(dt)
	}

	if elapsed < tilTabulate {
		dt := tilTabulate - elapsed
		fmt.Printf("waiting %v for tabulate\n", dt)
		tq <- e.c.After(dt)
	}

	dt := til
	fmt.Printf("waiting %v for next epoch\n", dt)

	tq <- e.c.After(dt)

	for ch := range tq {
		select {
		case <-e.ctx.Done():
			return
		case <-ch:
			sm.Advance()
		}
	}
}
