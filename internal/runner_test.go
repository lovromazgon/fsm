package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lovromazgon/fsm/example"
)

func TestFooFSM(t *testing.T) {
	ins := New[example.FooState, example.FooObservation](&example.FooFSM{LastState: "INITIAL"})
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
		if ins.Current().Done() {
			break
		}
	}

	if ins.Current().Failed() {
		t.Fatalf("failed with state %s", ins.Current())
	}
}
