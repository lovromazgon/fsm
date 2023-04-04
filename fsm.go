package fsm

import (
	"context"
)

type Definition[S comparable, O any] interface {
	States() []S
	Transitions() []Transition[S, O]
	New() Instance[S, O]
}

type Instance[S comparable, O any] interface {
	Observe(context.Context, FSM[S]) (O, error)
	Transition(context.Context, FSM[S], Transition[S, O], O) error
	Action(context.Context, FSM[S], O) error
}

type Transition[S comparable, O any] struct {
	From      S
	To        S
	Condition func(O) bool
}

type FSM[S comparable] interface {
	// Current returns the current state of the FSM instance.
	Current() S
	// Tick triggers next instance tick.
	Tick(ctx context.Context) error
}

func Instantiate[S comparable, O any](def Definition[S, O], runner func(Definition[S, O]) FSM[S]) FSM[S] {
	return runner(def)
}
