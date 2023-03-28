package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/lovromazgon/fsm"
	"github.com/lovromazgon/fsm/example"
)

func TestFooFSM(t *testing.T) {
	def := example.FooDef{}

	runner := fsm.Runner[example.FooState, example.FooEvent]{
		Definition:  def.Def(),
		Instantiate: Instantiate[example.FooState, example.FooEvent],
	}

	ins := runner.Run()
	fmt.Printf("%#v\n", ins)
	// fmt.Printf("%#v\n", ins.AvailableEvents())
	// fmt.Printf("can stop: %v\n", ins.Can(example.FooEventStop{}))

	fmt.Println("state:", ins.Current())
	fmt.Println("-------------------")

	for {
		err := ins.Tick(context.Background())
		fmt.Println("err:  ", err)
		fmt.Println("state:", ins.Current())
		fmt.Println("-------------------")
		if err != nil {
			break
		}
	}
}
