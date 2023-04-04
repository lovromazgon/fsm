package internal

import (
	"context"
	"fmt"
	"github.com/lovromazgon/fsm"
)

type FSM[S comparable, O any] struct {
	states      []S
	transitions []fsm.Transition[S, O]
	current     S
	instance    fsm.Instance[S, O]
}

var _ fsm.FSM[int] = &FSM[int, any]{}

func (i *FSM[S, O]) Current() S {
	return i.current
}

func (i *FSM[S, O]) Tick(ctx context.Context) error {
	o, err := i.instance.Observe(ctx, i)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	err = i.transition(ctx, o)
	if err != nil {
		return err
	}

	err = i.instance.Action(ctx, i, o)
	if err != nil {
		return err
	}

	// TODO persist instance
	return nil
}

func (i *FSM[S, O]) transition(ctx context.Context, o O) error {
	for _, t := range i.transitions {
		if t.From == i.current && t.Condition(o) {
			// found the transition we want to follow
			err := i.instance.Transition(ctx, i, t, o)
			if err != nil {
				return err
			}

			i.current = t.To
			return nil
		}
	}
	return nil // no applicable transition found, that's fine
}

func Instantiate[S comparable, O any](def fsm.Definition[S, O]) fsm.FSM[S] {
	i := &FSM[S, O]{
		states:      def.States(),
		transitions: def.Transitions(),
		current:     def.States()[0],
		instance:    def.New(),
	}

	return i
}
