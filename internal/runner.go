package internal

import (
	"context"
	"fmt"

	"github.com/lovromazgon/fsm"
)

type FSM[S fsm.State, O any] struct {
	states      []S
	transitions []fsm.Transition[S, O]
	current     S
	instance    fsm.FSM[S, O]
}

func (f *FSM[S, O]) Current() S {
	return f.current
}

func (f *FSM[S, O]) Tick(ctx context.Context) error {
	o, err := f.instance.Observe(ctx, f)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	err = f.transition(ctx, o)
	if err != nil {
		return err
	}

	err = f.instance.Action(ctx, f, o)
	if err != nil {
		return err
	}

	return nil
}

func (f *FSM[S, O]) transition(ctx context.Context, o O) error {
	for _, t := range f.transitions {
		if t.From == f.current && t.Condition(o) {
			// found the transition we want to follow
			err := f.instance.Transition(ctx, f, t, o)
			if err != nil {
				return err
			}

			f.current = t.To
			return nil
		}
	}
	return nil // no applicable transition found, that's fine
}

func New[S fsm.State, O any](instance fsm.FSM[S, O]) *FSM[S, O] {
	return &FSM[S, O]{
		states:      instance.States(),
		transitions: instance.Transitions(),
		current:     instance.States()[0],
		instance:    instance,
	}
}
