package internal

import (
	"context"
	"errors"

	"github.com/lovromazgon/fsm"
)

type Instance[S comparable, E fsm.Comparable[E]] struct {
	states      []S
	events      []E
	transitions []fsm.Transition[S, E]
	current     S
}

type foo struct{}

func (foo) Equals(foo) bool { return true }

var _ fsm.Instance[int, foo] = &Instance[int, foo]{}

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
		if got.Equals(want) {
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
		if t.From == i.current && t.Event.Equals(e) {
			i.current = t.To
			return nil
		}
	}
	return errors.New("can't")
}

func Instantiate[S comparable, E fsm.Comparable[E]](definition fsm.Definition[S, E]) fsm.Instance[S, E] {
	return &Instance[S, E]{
		states:      definition.States(),
		events:      definition.Events(),
		transitions: definition.Transitions(),
		current:     definition.States()[0],
	}
}
