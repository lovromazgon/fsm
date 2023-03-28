package fsm

import (
	"context"
)

type Definition[S comparable, E any] interface {
	States() []S
	Events() []E
	Transitions() []Transition[S, E]
	New() Instance[S, E]
}

type Instance[S comparable, E any] interface {
	Observe(context.Context, FSM[S]) (E, error)
	Action(context.Context, FSM[S]) error
	Transition(context.Context, FSM[S], Transition[S, E]) error
}

type Transition[S comparable, E any] struct {
	From  S
	To    S
	Event E
}

type FSM[S comparable] interface {
	// Current returns the current state of the FSM instance.
	Current() S
	// Tick triggers next instance tick.
	Tick(ctx context.Context) error
}

// type Runner[S comparable, E any] func(Definition[S, E]) FSM[S, E]

type Runner[S comparable, E any] struct {
	Definition  Definition[S, E]
	Instantiate func(Definition[S, E]) FSM[S]
}

func (r *Runner[S, E]) Run() FSM[S] {
	return r.Instantiate(r.Definition)
}
