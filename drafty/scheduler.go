package main

import (
	"sync"

	"github.com/katzenpost/core/epochtime"
)

type scheduler struct {
	sync.WaitGroup

	log    *logging.Logger
	haltCh chan interface{}
}

func (sch *scheduler) halt() {
	close(sch.haltCh)
	sch.Wait()
	sch.ch.Close()
}

// vote duration is 15 minutes, comprised of two 7.5 minute sections:
// a. vote exchange
// b. tabulation and signature exchange
// as described in "Panoramix Mix Network Public Key Infrastructure Specification"
// Voting begins at T + 2 hours where T is the beginning of the current epoch
// AND ends at T + 2 hours + 15 min.
func (sch *scheduler) worker() {
	current, elapsed, till := epochtime.Now()
	timer := time.NewTimer(till)
	defer func() {
		timer.Stop()
		sch.Done()
	}()

	for {
		select {
		case <-timer.C:
			// XXX blah

		}
	}
}
