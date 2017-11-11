// scheduler.go - a dir-auth scheduler
// Copyright (C) 2017  David Stainton, Nick Owens
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package scheduler

import (
	"context"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/katzenpost/authority/constants"
	"github.com/katzenpost/authority/statemachine"
	"github.com/katzenpost/core/epochtime"
)

type EpochScheduler struct {
	ctx context.Context
	c   clockwork.Clock
	et  *epochtime.Clock
}

func New(ctx context.Context, c clockwork.Clock) *EpochScheduler {
	return &EpochScheduler{ctx, c, epochtime.New(c)}
}

func (e *EpochScheduler) Run(sm statemachine.StateMachine) {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			e.Runone(sm)
		}
	}
}

func (e *EpochScheduler) Runone(sm statemachine.StateMachine) {
	_, elapsed, til := e.et.Now()

	var tq []<-chan time.Time

	if elapsed < constants.TilExchange {
		dt := constants.TilExchange - elapsed
		tq = append(tq, e.c.After(dt))
	}

	if elapsed < constants.TilTabulate {
		dt := constants.TilTabulate - elapsed
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
