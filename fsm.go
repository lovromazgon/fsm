package fsm

import (
	"context"
)

type Definition[S comparable, E any] interface {
	States() []S
	Events() []E
	Transitions() []Transition[S, E]
}

type OnTransition[S comparable, E any] interface {
	OnTransition(context.Context, Instance[S, E], Transition[S, E]) error
}

type Transition[S comparable, E any] struct {
	From  S
	To    S
	Event E
}

type Instance[S comparable, E any] interface {
	// AvailableEvents returns a list of events available in the current state.
	AvailableEvents() []E
	// Can returns true if event can occur in the current state.
	Can(E) bool
	// Current returns the current state of the FSM instance.
	Current() S
	// Send initiates a state transition with the event.
	Send(context.Context, E) error
}

type Runner[S comparable, E any] struct {
	Definition  Definition[S, E]
	Instantiate func(Definition[S, E]) Instance[S, E]
}

func (r *Runner[S, E]) Run() Instance[S, E] {
	return r.Instantiate(r.Definition)
}
