package fsm

import (
	"context"
)

type FSM[S State, O any] interface {
	States() []S
	Transitions() []Transition[S, O]

	Observe(context.Context, Helper[S]) (O, error)
	Transition(context.Context, Helper[S], Transition[S, O], O) error
	Action(context.Context, Helper[S], O) error
}

type State interface {
	~string
	Done() bool
	Failed() bool
}

type Transition[S State, O any] struct {
	From      S
	To        S
	Condition func(O) bool
}

// type FSM[S State] interface {
// 	// Current returns the current state of the FSM instance.
// 	Current() S
// 	// Tick triggers next instance tick.
// 	Tick(ctx context.Context) error
// }

type Helper[S State] interface {
	// Current returns the current state of the FSM instance.
	Current() S
}
