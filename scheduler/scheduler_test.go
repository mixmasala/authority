package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/katzenpost/core/epochtime"

	"github.com/jonboulle/clockwork"
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

	sched := NewEpochScheduler(ctx, clock)
	fsm := &fakeStateMachine{t, 0, make(chan struct{})}

	go sched.Run(fsm)

	fsm.expect(false)

	// run through some epochs
	for i := 0; i < 10; i++ {
		clock.Advance(tilExchange / 2)
		fsm.expect(false)
		clock.Advance(tilExchange / 2)
		fsm.expect(true)
		clock.Advance((tilTabulate - tilExchange) / 2)
		fsm.expect(false)
		clock.Advance((tilTabulate - tilExchange) / 2)
		fsm.expect(true)
		clock.Advance((epochtime.Period - tilTabulate) / 2)
		fsm.expect(false)
		clock.Advance((epochtime.Period - tilTabulate) / 2)
		fsm.expect(true)
	}

	if fsm.calls != 30 {
		t.Fatalf("expected 30 calls to Advance, got %d", fsm.calls)
	}
}
