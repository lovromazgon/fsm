package main

import (
	"log"

	"github.com/lovromazgon/fsm/example"
	"github.com/lovromazgon/fsm/temporal"
)

func main() {
	err := temporal.RunWorker[example.FooState, example.FooObservation](&example.FooFSM{})
	if err != nil {
		log.Fatalln("Failed to run worker", err)
	}
}
