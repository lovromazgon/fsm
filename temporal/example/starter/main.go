package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lovromazgon/fsm/example"
	"github.com/lovromazgon/fsm/temporal"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	f := temporal.New[example.FooState, example.FooObservation, *example.FooInstance](c, example.FooDef{})

	ticker := time.Tick(time.Second)
	i := 0
	for range ticker {
		i++
		if i == 10 {
			break
		}
		fmt.Println(f.Current())
	}
}
