package looplab

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	looplab "github.com/looplab/fsm"
	"github.com/lovromazgon/fsm"
)

type FSM[S comparable, O any] struct {
	stateStringer stringer[S]

	transitions []fsm.Transition[S, O]
	instance    fsm.Instance[S, O]
	fsm         *looplab.FSM
}

var _ fsm.FSM[int] = &FSM[int, any]{}

func (f FSM[S, O]) Current() S {
	return f.stateStringer.ToType(f.fsm.Current())
}

func (f FSM[S, O]) Tick(ctx context.Context) error {
	o, err := f.instance.Observe(ctx, f)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	for _, t := range f.transitions {
		if t.From == f.stateStringer.ToType(f.fsm.Current()) && t.Condition(o) {
			// this triggers transition and state action
			return f.fsm.Event(ctx, buildEventName(t, f.stateStringer), o)
		}
	}

	// no transition found, send dummy event to trigger action
	err = f.fsm.Event(ctx, f.fsm.Current()+"::ACTION", o)
	if err != nil && !errors.Is(err, looplab.NoTransitionError{}) {
		return err
	}

	return nil
}

func Instantiate[S comparable, O any](def fsm.Definition[S, O]) fsm.FSM[S] {
	f := &FSM[S, O]{
		stateStringer: newStringer(def.States()),

		transitions: def.Transitions(),
		instance:    def.New(),
	}

	events := make([]looplab.EventDesc, 0, len(def.Transitions())+len(def.States()))
	callbacks := make(map[string]looplab.Callback)
	for _, t := range def.Transitions() {
		transition := t
		event := buildEventName(t, f.stateStringer)
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{f.stateStringer.ToString(t.From)},
			Dst:  f.stateStringer.ToString(t.To),
		})
		callbacks["before_"+event] = func(ctx context.Context, e *looplab.Event) {
			err := f.instance.Transition(ctx, f, transition, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	for _, s := range def.States() {
		state := f.stateStringer.ToString(s)
		// add dummy events that will be used to execute action in case no
		// transition happens
		event := state + "::ACTION"
		events = append(events, looplab.EventDesc{
			Name: event,
			Src:  []string{state},
			Dst:  state,
		})
		callbacks["enter_"+state] = func(ctx context.Context, e *looplab.Event) {
			err := f.instance.Action(ctx, f, e.Args[0].(O))
			if err != nil {
				e.Cancel(err)
			}
		}
	}

	f.fsm = looplab.NewFSM(
		f.stateStringer.ToString(def.States()[0]),
		events,
		callbacks,
	)

	return f
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

func buildEventName[S comparable, O any](t fsm.Transition[S, O], stringer stringer[S]) string {
	return stringer.ToString(t.From) + "::" + stringer.ToString(t.To)
}
