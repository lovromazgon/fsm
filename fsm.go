package fsm

import (
	"context"
	"fmt"
)

type Definition[S comparable, E Comparable[E]] interface {
	States() []S
	Events() []E
	Transitions() []Transition[S, E]
}

type OnTransition[S comparable, E Comparable[E]] interface {
	OnTransition(context.Context, Instance[S, E], Transition[S, E]) error
}

type Transition[S comparable, E Comparable[E]] struct {
	From  S
	To    S
	Event E
}

func Print[S comparable, E Comparable[E]](def Definition[S, E]) {
	fmt.Println("STATES:")
	for _, s := range def.States() {
		fmt.Println("- ", s)
	}
	fmt.Println()
	fmt.Println("EVENTS:")
	for _, e := range def.Events() {
		fmt.Println("- ", e)
	}
	fmt.Println()
	fmt.Println("TRANSITIONS:")
	for _, t := range def.Transitions() {
		fmt.Println("- ", t)
	}
}

type Instance[S comparable, E Comparable[E]] interface {
	// AvailableEvents returns a list of events available in the current state.
	AvailableEvents() []E
	// Can returns true if event can occur in the current state.
	Can(E) bool
	// Current returns the current state of the FSM instance.
	Current() S
	// Send initiates a state transition with the event.
	Send(context.Context, E) error
}

type Runner[S comparable, E Comparable[E]] struct {
	Definition  Definition[S, E]
	Instantiate func(Definition[S, E]) Instance[S, E]
}

func (r *Runner[S, E]) Run() Instance[S, E] {
	return r.Instantiate(r.Definition)
}

type Comparable[T any] interface {
	Equals(T) bool
}
