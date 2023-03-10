package example

import (
	"context"
	"fmt"

	"github.com/lovromazgon/fsm"
)

// FSM definition
type FooFSM struct{}
type FooState string
type FooEvent interface{ fooEvent() }

func (f FooFSM) FSMDefinition() fsm.Definition[FooState, FooEvent] {
	return f
}

func (f FooFSM) BeforeTransition(ctx context.Context, i fsm.Instance[FooState, FooEvent], t fsm.Transition[FooState, FooEvent]) error {
	fmt.Printf("BEFORE: currently %v, going to %v, because of %T\n", i.Current(), t.To, t.Event)
	return nil
}

func (f FooFSM) AfterTransition(ctx context.Context, i fsm.Instance[FooState, FooEvent], t fsm.Transition[FooState, FooEvent]) error {
	fmt.Printf("AFTER: currently %v, came from %v, because of %T\n", i.Current(), t.From, t.Event)
	fmt.Printf("AFTER: full event: %+v\n", t.Event)
	return nil
}

func (FooFSM) States() []FooState {
	return []FooState{
		FooStateRunning,
		FooStateWaiting,
		FooStateDone,
		FooStateFailed,
	}
}

func (FooFSM) Events() []FooEvent {
	return []FooEvent{
		FooEventWait{},
		FooEventStop{},
		FooEventFail{},
	}
}

// Transitions returns the possible transitions.
func (FooFSM) Transitions() []fsm.Transition[FooState, FooEvent] {
	return []fsm.Transition[FooState, FooEvent]{
		{Event: FooEventWait{}, From: FooStateRunning, To: FooStateWaiting},
		{Event: FooEventStop{}, From: FooStateWaiting, To: FooStateDone},
		{Event: FooEventFail{}, From: FooStateWaiting, To: FooStateFailed},
		{Event: FooEventFail{}, From: FooStateRunning, To: FooStateFailed},
		{Event: FooEventFail{}, From: FooStateDone, To: FooStateFailed},
	}
}

// define states
const (
	FooStateRunning FooState = "Running"
	FooStateWaiting FooState = "Waiting"
	FooStateDone    FooState = "Done"
	FooStateFailed  FooState = "Failed"
)

// define events
type (
	FooEventWait struct{}
	FooEventStop struct {
		EventsCanHaveFields int
	}
	FooEventFail struct {
		Err error
	}
)

func (FooEventWait) fooEvent() {}
func (FooEventStop) fooEvent() {}
func (FooEventFail) fooEvent() {}
