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
	instance    fsm.FSM[S, O]
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

func New[S fsm.State, O any](ins fsm.FSM[S, O]) *FSM[S, O] {
	var f *FSM[S, O]

	events := make([]looplab.EventDesc, 0, len(ins.Transitions())+len(ins.States()))
	callbacks := make(map[string]looplab.Callback)
	for _, t := range ins.Transitions() {
		transition := t
		event := eventNameForTransition(t)
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{string(t.From)},
			Dst:  string(t.To),
		})
		callbacks["before_"+event] = func(ctx context.Context, e *looplab.Event) {
			err := ins.Transition(ctx, f, transition, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	for _, s := range ins.States() {
		// add dummy events that will be used to execute action in case no
		// transition happens
		event := eventNameForState(s)
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{string(s)},
			Dst:  string(s),
		})
		callbacks["after_"+event] = func(ctx context.Context, e *looplab.Event) {
			err := ins.Action(ctx, f, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	f = &FSM[S, O]{
		transitions: ins.Transitions(),
		instance:    ins,
		fsm: looplab.NewFSM(
			string(ins.States()[0]),
			events,
			callbacks,
		),
	}
	return f
}

func eventNameForTransition[S fsm.State, O any](t fsm.Transition[S, O]) string {
	return string(t.From) + "::" + string(t.To)
}

func eventNameForState[S ~string](s S) string {
	return string(s) + "::action"
}
