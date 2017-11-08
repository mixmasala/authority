package authority

import (
)

type AuthorityState int

const (
	StateInvalid AuthorityState = iota
	StateWait
	StateExchange
	StateTabluate

)

type StateMachine interface {
	Advance()
}
