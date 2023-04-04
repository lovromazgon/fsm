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

// type Runner[S comparable, E any] func(Definition[S, E]) FSM[S, E]

type Runner[S comparable, O any] struct {
	Definition  Definition[S, O]
	Instantiate func(Definition[S, O]) FSM[S]
}

func (r *Runner[S, E]) Run() FSM[S] {
	return r.Instantiate(r.Definition)
}
