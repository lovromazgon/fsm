package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lovromazgon/fsm"
)

func TestFooFSM(t *testing.T) {
	def := FooFSM{}
	// fsm.Print(def.FSMDefinition())

	runner := fsm.Runner[FooState, FooEvent]{
		Definition:  def.FSMDefinition(),
		Instantiate: Instantiate[FooState, FooEvent],
	}

	ins := runner.Run()
	fmt.Printf("%#v\n", ins)
	fmt.Printf("%#v\n", ins.AvailableEvents())
	fmt.Printf("%#v\n", ins.Can(FooEventStop{}))

	fmt.Println(ins.Current())
	fmt.Println("-------------------")

	err := ins.Send(context.Background(), FooEventStop{})
	fmt.Println(err)
	fmt.Println(ins.Current())
	fmt.Println("-------------------")

	err = ins.Send(context.Background(), FooEventFail{Err: errors.New("whoops")})
	fmt.Println(err)
	fmt.Println(ins.Current())
	fmt.Println("-------------------")
}

// FSM definition
type FooFSM struct{}

func (f FooFSM) FSMDefinition() fsm.Definition[FooState, FooEvent] {
	return f
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

type FooState string
type FooEvent interface{ fooEvent() }

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
