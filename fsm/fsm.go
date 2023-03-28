package fsm

import (
	"context"
)

type FSM interface {
	StateFunctions() map[string]StateFunction
	Observe(context.Context, H) (string, error)
}

type StateFunction func(context.Context, H) error

type H interface {
	State() string
}
