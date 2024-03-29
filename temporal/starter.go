package temporal

import (
	"context"
	"log"
	"reflect"

	"github.com/lovromazgon/fsm"
	"go.temporal.io/sdk/client"
)

type FSM[S fsm.State] struct {
	c  client.Client
	wr client.WorkflowRun
}

func (f FSM[S]) Current() S {
	ev, err := f.c.QueryWorkflow(context.Background(), f.wr.GetID(), f.wr.GetRunID(), "state", nil)
	if err != nil {
		panic(err)
	}
	var s S
	err = ev.Get(&s)
	if err != nil {
		panic(err)
	}
	return s
}

func (f FSM[S]) Tick(ctx context.Context) error {
	return f.c.SignalWorkflow(ctx, f.wr.GetID(), f.wr.GetRunID(), "tick", nil)
}

func New[S fsm.State, O any](c client.Client, ins fsm.FSM[S, O]) *FSM[S] {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "fsm",
	}

	wr, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflowNameForFSM(ins), ins)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	return &FSM[S]{
		c:  c,
		wr: wr,
	}
}

func workflowNameForFSM(def any) string {
	return reflect.TypeOf(def).String()
}
