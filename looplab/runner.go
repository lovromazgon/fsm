package looplab

import (
	"context"
	"errors"
	"fmt"

	looplab "github.com/looplab/fsm"
	"github.com/lovromazgon/fsm"
)

type FSM[S fsm.State, O any] struct {
	transitions []fsm.Transition[S, O]
	instance    fsm.Instance[S, O]
	fsm         *looplab.FSM
}

func (f FSM[S, O]) Current() S {
	return S(f.fsm.Current())
}

func (f FSM[S, O]) Tick(ctx context.Context) error {
	o, err := f.instance.Observe(ctx, f)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	for _, t := range f.transitions {
		if t.From == S(f.fsm.Current()) && t.Condition(o) {
			// this triggers transition
			err := f.fsm.Event(ctx, eventNameForTransition(t), o)
			if err != nil {
				return err
			}
			break
		}
	}

	// send another dummy event to trigger action
	err = f.fsm.Event(ctx, eventNameForState(f.fsm.Current()), o)
	if err != nil && !errors.Is(err, looplab.NoTransitionError{}) {
		return err
	}

	return nil
}

func New[S fsm.State, O any, I fsm.Instance[S, O]](def fsm.Definition[S, O, I]) fsm.FSM[S] {
	f := &FSM[S, O]{
		transitions: def.Transitions(),
		instance:    def.New(),
	}

	events := make([]looplab.EventDesc, 0, len(def.Transitions())+len(def.States()))
	callbacks := make(map[string]looplab.Callback)
	for _, t := range def.Transitions() {
		transition := t
		event := eventNameForTransition(t)
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{string(t.From)},
			Dst:  string(t.To),
		})
		callbacks["before_"+event] = func(ctx context.Context, e *looplab.Event) {
			err := f.instance.Transition(ctx, f, transition, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	for _, s := range def.States() {
		// add dummy events that will be used to execute action in case no
		// transition happens
		event := eventNameForState(s)
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{string(s)},
			Dst:  string(s),
		})
		callbacks["after_"+event] = func(ctx context.Context, e *looplab.Event) {
			err := f.instance.Action(ctx, f, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	f.fsm = looplab.NewFSM(
		string(def.States()[0]),
		events,
		callbacks,
	)

	return f
}

func eventNameForTransition[S fsm.State, O any](t fsm.Transition[S, O]) string {
	return string(t.From) + "::" + string(t.To)
}

func eventNameForState[S ~string](s S) string {
	return string(s) + "::action"
}

// check that FSM implements fsm.FSM, use dummyState as type fsm.State.
var _ fsm.FSM[dummyState] = &FSM[dummyState, any]{}

type dummyState string

func (dummyState) Done() bool   { return false }
func (dummyState) Failed() bool { return false }
