// Package example contains multiple structs that together form a runnable state
// machine. The structs and their functionality:
//   - FooFSM - represents a single instance of a running state machine. It
//     contains methods for observing the state, for reacting to a transition
//     change and for executing an action in a specific state. It can store any
//     state in its struct fields.
//   - FooObservation - holds the data observed in FooFSM.Observe. Based on
//     the observation the FSM chooses the transition to a new state (if any).
//   - FooState - represents a state of the state machine.
package example

import (
	"context"
	"fmt"

	"github.com/lovromazgon/fsm"
)

// FooFSM is the Foo state machine.
type FooFSM struct {
	LastState FooState
}

func (*FooFSM) States() []FooState {
	return []FooState{
		FooStateRunning,
		FooStateWaiting,
		FooStateDone,
		FooStateFailed,
	}
}

// Transitions returns the possible transitions. Each transition knows its own
// conditions when it should apply. Transitions are ranked by priority, if two
// transitions would apply based on an observation, the first one takes
// precedence.
func (*FooFSM) Transitions() []fsm.Transition[FooState, FooObservation] {
	return []fsm.Transition[FooState, FooObservation]{
		{From: FooStateRunning, To: FooStateWaiting, Condition: func(o FooObservation) bool {
			return o.SomethingToObserve == "wait"
		}},
		{From: FooStateWaiting, To: FooStateDone, Condition: func(o FooObservation) bool {
			return o.SomethingToObserve == "done"
		}},
		{From: FooStateWaiting, To: FooStateFailed, Condition: func(o FooObservation) bool {
			return !o.ServiceIsUp
		}},
		{From: FooStateRunning, To: FooStateFailed, Condition: func(o FooObservation) bool {
			return !o.ServiceIsUp
		}},
		{From: FooStateDone, To: FooStateFailed, Condition: func(o FooObservation) bool {
			return !o.ServiceIsUp
		}},
	}
}

func (a *FooFSM) Observe(ctx context.Context, i fsm.Helper[FooState]) (FooObservation, error) {
	defer func() {
		a.LastState = i.Current() // store last state after observation
	}()
	if a.LastState != i.Current() {
		fmt.Println("new state, let's just execute action")
		return FooObservation{ServiceIsUp: true}, nil
	}
	fmt.Println("same state as before, transitioning")
	switch i.Current() {
	case FooStateRunning:
		return FooObservation{ServiceIsUp: true, SomethingToObserve: "wait"}, nil
	case FooStateWaiting:
		return FooObservation{ServiceIsUp: true, SomethingToObserve: "done"}, nil
	default:
		return FooObservation{ServiceIsUp: false}, nil
	}
}

func (a *FooFSM) Transition(ctx context.Context, i fsm.Helper[FooState], t fsm.Transition[FooState, FooObservation], o FooObservation) error {
	fmt.Printf("BEFORE: currently %v, going to %v\n", i.Current(), t.To)
	fmt.Printf("BEFORE: observation: %+v\n", o)
	return nil
}

func (a *FooFSM) Action(ctx context.Context, i fsm.Helper[FooState], o FooObservation) error {
	fmt.Printf("ACTION: currently %v, old %v\n", i.Current(), a.LastState)
	fmt.Printf("ACTION: observation: %+v\n", o)
	return nil
}

// FooState is a state in the FooFSM state machine.
type FooState string

const (
	FooStateRunning FooState = "Running"
	FooStateWaiting FooState = "Waiting"
	FooStateDone    FooState = "Done"
	FooStateFailed  FooState = "Failed"
)

func (s FooState) Done() bool {
	return s == FooStateDone || s == FooStateFailed
}

func (s FooState) Failed() bool {
	return s == FooStateFailed
}

// FooObservation is the observation returned by FooFSM.Observe.
type FooObservation struct {
	SomethingToObserve string
	ServiceIsUp        bool
}
