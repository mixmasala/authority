// statemachine.go - FSM for dir-auth voting server
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

package statemachine

import (
	"errors"
	"time"

	"github.com/katzenpost/authority/constants"
)

// AuthorityState is a type for expressing
// the current state of the dir-auth FSM
type AuthorityState int

const (
	// StateInvalid is an invalid state
	StateInvalid AuthorityState = iota
	// StateWait is used to denote the phase
	// where the Dir-Auth server is waiting for
	// the next voting period to begin
	StateWait
	// StateExchange is used to denote the voting exchange
	// phase of a given epoch
	StateExchange
	// StateTabulate is used to denote the tabulation and
	// signature exchange phase of the given epoch
	StateTabulate
)

// StateMachine is an interface for implementing
// a statemachine for changing the Directory Authority
// schedule
type StateMachine interface {
	Advance()
}

// SimpleStateMachine implements the StateMachine interface
type SimpleStateMachine struct {
	State            AuthorityState
	VoteHandler      func()
	SignatureHandler func()
}

// Advance advanced our super simple branch-less state machine
func (s *SimpleStateMachine) Advance() {
	switch s.State {
	case StateWait:
		s.State = StateExchange
		s.VoteHandler()
	case StateExchange:
		s.State = StateTabulate
		s.SignatureHandler()
	case StateTabulate:
		s.State = StateWait
	}
}

// New returns a new SimpleStateMachine or an error
func New(elapsed time.Duration, voteHandler func(), signatureHandler func()) (*SimpleStateMachine, error) {
	if elapsed > constants.EpochDuration {
		return nil, errors.New("elapsed time cannot exceed epoch duration")
	}
	s := SimpleStateMachine{}
	s.VoteHandler = voteHandler
	s.SignatureHandler = signatureHandler
	if elapsed < constants.TilExchange {
		s.State = StateWait
		return &s, nil
	}
	if elapsed > constants.TilExchange && elapsed < constants.TilTabulate {
		s.State = StateExchange
		return &s, nil
	}
	if elapsed > constants.TilTabulate && elapsed < constants.TilEndOfTabulate {
		s.State = StateTabulate
		return &s, nil
	}
	if elapsed > constants.TilEndOfTabulate {
		s.State = StateWait
		return &s, nil
	}
	return nil, errors.New("wtf")
}
