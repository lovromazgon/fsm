package main

import (
	"github.com/lovromazgon/fsm/fsm/example"
	"github.com/lovromazgon/fsm/fsm/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	fsm := &example.FooFSM{}
	w := worker.New(c, "fsm", worker.Options{})
	temporal.RegisterFSMWorkflow(fsm, w, w)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
