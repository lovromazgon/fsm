package example

import (
	"context"
	"fmt"
	"github.com/lovromazgon/fsm"
)

// FooDef is the definition of a FSM.
type FooDef struct{}
type FooState string
type FooEvent interface{ fooEvent() }

func (f FooDef) Def() fsm.Definition[FooState, FooEvent] {
	return f
}

func (FooDef) New() fsm.Instance[FooState, FooEvent] {
	return &FooInstance{}
}

func (FooDef) States() []FooState {
	return []FooState{
		FooStateRunning,
		FooStateWaiting,
		FooStateDone,
		FooStateFailed,
	}
}

func (FooDef) Events() []FooEvent {
	return []FooEvent{
		FooEventWait{},
		FooEventStop{},
		FooEventFail{},
	}
}

// Transitions returns the possible transitions.
func (FooDef) Transitions() []fsm.Transition[FooState, FooEvent] {
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

type FooInstance struct {
	LastState FooState
}

func (a *FooInstance) Observe(ctx context.Context, i fsm.FSM[FooState]) (FooEvent, error) {
	defer func() {
		a.LastState = i.Current()
	}()
	if a.LastState != i.Current() {
		fmt.Println("new state, let's just execute action")
		return nil, nil
	}
	fmt.Println("same state as before, transitioning")
	switch i.Current() {
	case FooStateRunning:
		return FooEventWait{}, nil
	case FooStateWaiting:
		return FooEventStop{}, nil
	default:
		return FooEventFail{}, nil
	}
}

func (a *FooInstance) Action(ctx context.Context, i fsm.FSM[FooState]) error {
	fmt.Printf("ACTION: currently %v, old %v\n", i.Current(), a.LastState)
	return nil
}

func (a *FooInstance) Transition(ctx context.Context, i fsm.FSM[FooState], t fsm.Transition[FooState, FooEvent]) error {
	fmt.Printf("BEFORE: currently %v, going to %v, because of %T\n", i.Current(), t.To, t.Event)
	fmt.Printf("BEFORE: full event: %+v\n", t.Event)
	return nil
}
