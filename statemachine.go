package authority

import ()

type AuthorityState int

const (
	StateInvalid AuthorityState = iota
	StateWait
	StateExchange
	StateTabulate
)

type StateMachine interface {
	Advance()
}

type SimpleStateMachine struct {
	state AuthorityState
}

func (s *SimpleStateMachine) Advance() {
	// XXX fix me
	if s.state == StateTabulate {
		s.state = StateWait
	} else {
		s.state += 1
	}
}
