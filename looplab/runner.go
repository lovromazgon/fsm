package looplab

import (
	"context"
	"reflect"
	"sync"

	looplab "github.com/looplab/fsm"
	"github.com/lovromazgon/fsm"
)

type Instance[S comparable, E any] struct {
	stateStringer stringer[S]
	eventStringer stringer[E]

	fsm *looplab.FSM
}

var _ fsm.Instance[int, any] = &Instance[int, any]{}

func (i *Instance[S, E]) AvailableEvents() []E {
	events := i.fsm.AvailableTransitions()
	es := make([]E, len(events))
	for j, event := range events {
		es[j] = i.eventStringer.ToType(event)
	}
	return es
}

func (i *Instance[S, E]) Can(want E) bool {
	return i.fsm.Can(i.eventStringer.ToString(want))
}

func (i *Instance[S, E]) Current() S {
	return i.stateStringer.ToType(i.fsm.Current())
}

func (i *Instance[S, E]) Send(ctx context.Context, e E) error {
	return i.fsm.Event(ctx, i.eventStringer.ToString(e), e)
}

func Instantiate[S comparable, E any](def fsm.Definition[S, E]) fsm.Instance[S, E] {
	i := &Instance[S, E]{
		stateStringer: newStringer(def.States()),
		eventStringer: newStringer(def.Events()),
	}

	events := make([]looplab.EventDesc, len(def.Transitions()))
	for j, t := range def.Transitions() {
		events[j] = looplab.EventDesc{
			Name: i.eventStringer.ToString(t.Event),
			Src:  []string{i.stateStringer.ToString(t.From)},
			Dst:  i.stateStringer.ToString(t.To),
		}
	}

	callbacks := make(map[string]looplab.Callback)
	if ot, ok := def.(fsm.BeforeTransition[S, E]); ok {
		callbacks["before_event"] = func(ctx context.Context, event *looplab.Event) {
			e := event.Args[0].(E)
			transition := fsm.Transition[S, E]{
				From:  i.stateStringer.ToType(event.Src),
				To:    i.stateStringer.ToType(event.Dst),
				Event: e,
			}

			err := ot.BeforeTransition(ctx, i, transition)
			if err != nil {
				event.Cancel(err)
			}
		}
	}
	if at, ok := def.(fsm.AfterTransition[S, E]); ok {
		callbacks["after_event"] = func(ctx context.Context, event *looplab.Event) {
			e := event.Args[0].(E)
			transition := fsm.Transition[S, E]{
				From:  i.stateStringer.ToType(event.Src),
				To:    i.stateStringer.ToType(event.Dst),
				Event: e,
			}

			err := at.AfterTransition(ctx, i, transition)
			if err != nil {
				event.Err = err
			}
		}
	}

	i.fsm = looplab.NewFSM(
		i.stateStringer.ToString(def.States()[0]),
		events,
		callbacks,
	)

	return i
}

// cachedStringers caches stringers, so they are not recreated every time.
var cachedStringers = sync.Map{}

func newStringer[T any](list []T) stringer[T] {
	t := reflect.TypeOf(new(T)).Elem()
	if s, ok := cachedStringers.Load(t); ok {
		// take existing stringer from cache
		return s.(stringer[T])
	}

	var toStringFunc func(T) string
	switch t.Kind() {
	case reflect.Interface:
		toStringFunc = func(t T) string {
			return reflect.TypeOf(t).String()
		}
	default:
		toStringFunc = func(t T) string {
			return reflect.ValueOf(t).String()
		}
	}

	mapping := make(map[string]T)
	for _, t := range list {
		mapping[toStringFunc(t)] = t
	}

	s := stringer[T]{
		mapping:      mapping,
		toStringFunc: toStringFunc,
	}

	cachedStringers.Store(t, s)
	return s
}

type stringer[T any] struct {
	mapping      map[string]T
	toStringFunc func(t T) string
}

func (c stringer[T]) ToString(t T) string {
	return c.toStringFunc(t)
}

func (c stringer[T]) ToType(s string) T {
	return c.mapping[s]
}
