package main

import (
	"context"
	"go.temporal.io/sdk/client"
	"log"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "fsm_workflowID",
		TaskQueue: "fsm",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "foofsm")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	err = we.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow done.")
}
