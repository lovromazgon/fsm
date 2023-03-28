package internal

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/lovromazgon/fsm"
)

type FSM[S comparable, E any] struct {
	states      []S
	events      []E
	transitions []fsm.Transition[S, E]
	current     S
	instance    fsm.Instance[S, E]

	eventEquals func(E, E) bool
}

var _ fsm.FSM[int] = &FSM[int, any]{}

func (i *FSM[S, E]) AvailableEvents() []E {
	var events []E
	for _, t := range i.transitions {
		if t.From == i.current {
			events = append(events, t.Event)
		}
	}
	return events
}

func (i *FSM[S, E]) Can(want E) bool {
	for _, got := range i.AvailableEvents() {
		if i.eventEquals(got, want) {
			return true
		}
	}
	return false
}

func (i *FSM[S, E]) Current() S {
	return i.current
}

func (i *FSM[S, E]) Tick(ctx context.Context) error {
	e, err := i.instance.Observe(ctx, i)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	switch {
	case reflect.ValueOf(e).IsValid():
		err = i.transition(ctx, e)
	default:
		err = i.action(ctx)
	}
	if err != nil {
		return err
	}

	// TODO persist instance
	return nil
}

func (i *FSM[S, E]) action(ctx context.Context) error {
	return i.instance.Action(ctx, i)
}

func (i *FSM[S, E]) transition(ctx context.Context, e E) error {
	var transition fsm.Transition[S, E]
	var found bool
	for _, t := range i.transitions {
		if t.From == i.current && i.eventEquals(t.Event, e) {
			transition = t
			found = true
			break
		}
	}
	if !found {
		return errors.New("can't")
	}

	transition.Event = e // overwrite event so we can send it to callback

	err := i.instance.Transition(ctx, i, transition)
	if err != nil {
		return err
	}

	i.current = transition.To
	return nil
}

func (i *FSM[S, E]) init() {
	e := new(E)

	switch reflect.TypeOf(e).Elem().Kind() {
	case reflect.Interface, reflect.Pointer:
		i.eventEquals = func(e1, e2 E) bool {
			return reflect.TypeOf(e1).String() == reflect.TypeOf(e2).String()
		}
	default:
		i.eventEquals = func(e1, e2 E) bool {
			return reflect.ValueOf(e1).Interface() == reflect.ValueOf(e2).Interface()
		}
	}
}

func Instantiate[S comparable, E any](def fsm.Definition[S, E]) fsm.FSM[S] {
	i := &FSM[S, E]{
		states:      def.States(),
		events:      def.Events(),
		transitions: def.Transitions(),
		current:     def.States()[0],
		instance:    def.New(),
	}
	i.init()

	return i
}
