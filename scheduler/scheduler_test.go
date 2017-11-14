package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/katzenpost/authority/constants"
	"github.com/katzenpost/core/epochtime"
)

type fakeStateMachine struct {
	t       *testing.T
	calls   uint32
	advance chan struct{}
}

func (f *fakeStateMachine) Advance() {
	atomic.AddUint32(&f.calls, 1)
	f.advance <- struct{}{}
}

func (f *fakeStateMachine) expect(want bool) {
	arrived := false
	select {
	case <-f.advance:
		arrived = true
	case <-time.After(1 * time.Millisecond):
	}

	if want != arrived {
		f.t.Logf("want=%v arrived=%v calls=%d", want, arrived, f.calls)
		if want == true {
			f.t.Fatal("did not receive expected advance")
		} else {
			f.t.Fatalf("unexpected advance, calls=%d", f.calls)
		}
	}
}

func TestSchedulerRun(t *testing.T) {
	clock := clockwork.NewFakeClockAt(epochtime.Epoch)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := New(ctx, clock)
	fsm := &fakeStateMachine{t, 0, make(chan struct{})}

	go sched.Run(fsm)

	fsm.expect(false)

	// run through some epochs
	for i := 0; i < 10; i++ {
		clock.Advance(constants.TilExchange / 2)
		fsm.expect(false)
		clock.Advance(constants.TilExchange / 2)
		fsm.expect(true)
		clock.Advance((constants.TilTabulate - constants.TilExchange) / 2)
		fsm.expect(false)
		clock.Advance((constants.TilTabulate - constants.TilExchange) / 2)
		fsm.expect(true)
		clock.Advance((epochtime.Period - constants.TilTabulate) / 2)
		fsm.expect(false)
		clock.Advance((epochtime.Period - constants.TilTabulate) / 2)
		fsm.expect(true)
	}

	if fsm.calls != 30 {
		t.Fatalf("expected 30 calls to Advance, got %d", fsm.calls)
	}
}
