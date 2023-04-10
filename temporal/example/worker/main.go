package main

import (
	"github.com/lovromazgon/fsm/example"
	"github.com/lovromazgon/fsm/temporal"
)

func main() {
	temporal.RunWorker[example.FooState, example.FooObservation, *example.FooInstance](example.FooDef{})
}
