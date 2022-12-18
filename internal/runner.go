package internal

import (
	"context"
	"errors"
	"reflect"

	"github.com/lovromazgon/fsm"
)

type Instance[S comparable, E any] struct {
	states      []S
	events      []E
	transitions []fsm.Transition[S, E]
	current     S

	beforeTransition fsm.BeforeTransition[S, E]
	afterTransition  fsm.AfterTransition[S, E]

	eventEquals func(E, E) bool
}

var _ fsm.Instance[int, any] = &Instance[int, any]{}

func (i *Instance[S, E]) AvailableEvents() []E {
	var events []E
	for _, t := range i.transitions {
		if t.From == i.current {
			events = append(events, t.Event)
		}
	}
	return events
}

func (i *Instance[S, E]) Can(want E) bool {
	for _, got := range i.AvailableEvents() {
		if i.eventEquals(got, want) {
			return true
		}
	}
	return false
}

func (i *Instance[S, E]) Current() S {
	return i.current
}

func (i *Instance[S, E]) Send(ctx context.Context, e E) error {
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
	if i.beforeTransition != nil {
		err := i.beforeTransition.BeforeTransition(ctx, i, transition)
		if err != nil {
			return err
		}
	}
	i.current = transition.To
	if i.afterTransition != nil {
		err := i.afterTransition.AfterTransition(ctx, i, transition)
		if err != nil {
			return err
		}
	}

	return nil

}

func (i *Instance[S, E]) init() {
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

func Instantiate[S comparable, E any](def fsm.Definition[S, E]) fsm.Instance[S, E] {
	i := &Instance[S, E]{
		states:      def.States(),
		events:      def.Events(),
		transitions: def.Transitions(),
		current:     def.States()[0],
	}
	i.init()

	if ot, ok := def.(fsm.BeforeTransition[S, E]); ok {
		i.beforeTransition = ot
	}
	if at, ok := def.(fsm.AfterTransition[S, E]); ok {
		i.afterTransition = at
	}

	return i
}
