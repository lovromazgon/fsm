package fsm

import (
	"context"
	"fmt"
)

type FSMDefinition[S any, E any] interface {
	States() []S
	Events() []E
	Transitions() []Transition[S, E]
}

type OnTransition[S any, E any] interface {
	OnTransition(context.Context, Transition[S, E]) error
}

type BeforeEvent[S any, E any] interface {
	BeforeEvent(context.Context, Transition[S, E]) error
}

type AfterEvent[S any, E any] interface {
	AfterEvent(context.Context, Transition[S, E]) error
}

type Transition[S any, E any] struct {
	From  S
	To    S
	Event E
}

func Print[S any, E any](def FSMDefinition[S, E]) {
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
