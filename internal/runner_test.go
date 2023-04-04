package internal

import (
	"context"
	"fmt"
	"github.com/lovromazgon/fsm"
	"github.com/lovromazgon/fsm/example"
	"testing"
	"time"
)

func TestFooFSM(t *testing.T) {
	def := example.FooDef{}

	runner := fsm.Runner[example.FooState, example.FooObservation]{
		Definition:  def,
		Instantiate: Instantiate[example.FooState, example.FooObservation],
	}

	ins := runner.Run()
	fmt.Printf("%#v\n", ins)

	fmt.Println("state:", ins.Current())
	fmt.Println("-------------------")

	for {
		// the sleep simulates delay between FSM ticks
		time.Sleep(time.Second / 2)

		err := ins.Tick(context.Background())
		fmt.Println("err:  ", err)
		fmt.Println("state:", ins.Current())
		fmt.Println("-------------------")
		if err != nil {
			t.Fatal(err)
		}
		if ins.Current() == example.FooStateFailed {
			break
		}
	}
}
