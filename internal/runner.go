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
	for _, t := range i.transitions {
		if t.From == i.current && i.eventEquals(t.Event, e) {
			i.current = t.To
			return nil
		}
	}
	return errors.New("can't")
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

func Instantiate[S comparable, E any](definition fsm.Definition[S, E]) fsm.Instance[S, E] {
	i := &Instance[S, E]{
		states:      definition.States(),
		events:      definition.Events(),
		transitions: definition.Transitions(),
		current:     definition.States()[0],
	}
	i.init()
	return i
}
