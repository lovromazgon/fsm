package example

import (
	"context"
	"fmt"
	"github.com/lovromazgon/fsm/fsm"
)

type FooFSM struct {
	MyData string
}

const (
	FooStateRunning = "Running"
	FooStateWaiting = "Waiting"
	FooStateDone    = "Done"
	FooStateFailed  = "Failed"
)

func (a *FooFSM) StateFunctions() map[string]fsm.StateFunction {
	return map[string]fsm.StateFunction{
		FooStateRunning: a.Running,
		FooStateWaiting: a.Waiting,
		FooStateDone:    a.Done,
		FooStateFailed:  a.Failed,
	}
}

func (a *FooFSM) Observe(ctx context.Context, helper fsm.H) (string, error) {
	// TODO observe some external data
	for state := range a.StateFunctions() {
		return state, nil
	}
	panic("unreachable")
}

func (a *FooFSM) Running(ctx context.Context, helper fsm.H) error {
	a.log(ctx, helper)
	return nil
}

func (a *FooFSM) Waiting(ctx context.Context, helper fsm.H) error {
	a.log(ctx, helper)
	return nil
}

func (a *FooFSM) Done(ctx context.Context, helper fsm.H) error {
	a.log(ctx, helper)
	return nil
}

func (a *FooFSM) Failed(ctx context.Context, helper fsm.H) error {
	a.log(ctx, helper)
	return nil
}

func (a *FooFSM) log(ctx context.Context, helper fsm.H) {
	fmt.Println("=== I'M IN STATE:", helper.State())
}
