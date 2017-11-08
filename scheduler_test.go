package authority

import (
	"testing"
	"context"
	"sync/atomic"

	"github.com/katzenpost/core/epochtime"

	"github.com/jonboulle/clockwork"
)

type fakeStateMachine struct {
	t *testing.T
	calls uint32
	advance chan struct{}
}

func (f *fakeStateMachine) Advance() {
	f.t.Log("Advancing...")
	atomic.AddUint32(&f.calls, 1)
	f.advance <- struct{}{}
}

func TestSchedulerRun(t *testing.T) {
	clock := clockwork.NewFakeClockAt(epochtime.Epoch)

	sched := NewEpochScheduler(context.TODO(), clock)

	fsm := &fakeStateMachine{
		t,
		0,
		make(chan struct{}),
	}

	go sched.Run(fsm)

	t.Logf("T0=%s", clock.Now())
	clock.BlockUntil(1)
	clock.Advance(tilExchange+1)
	t.Logf("T1=%s", clock.Now())

	<-fsm.advance

	if fsm.calls != 1 {
		t.Fatalf("expected one advance call, got %d", fsm.calls)
	}

	clock.BlockUntil(1)
	clock.Advance(tilTabulate - tilExchange + 1)

	if fsm.calls != 2 {
		t.Fatalf("expected 2 advance calls, got %d", fsm.calls)
	}

	clock.BlockUntil(1)
	clock.Advance(epochtime.Period - tilTabulate + 1)

	if fsm.calls != 3 {
		t.Fatalf("expected 3 advance calls, got %d", fsm.calls)
	}
}

