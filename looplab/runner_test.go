package looplab

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lovromazgon/fsm"
	"github.com/lovromazgon/fsm/example"
)

func TestFooFSM(t *testing.T) {
	def := example.FooFSM{}

	runner := fsm.Runner[example.FooState, example.FooEvent]{
		Definition:  def.FSMDefinition(),
		Instantiate: Instantiate[example.FooState, example.FooEvent],
	}

	ins := runner.Run()
	fmt.Printf("%#v\n", ins)
	fmt.Printf("%#v\n", ins.AvailableEvents())
	fmt.Printf("can stop: %v\n", ins.Can(example.FooEventStop{}))

	fmt.Println("state:", ins.Current())
	fmt.Println("-------------------")

	err := ins.Send(context.Background(), example.FooEventStop{})
	fmt.Println("err:  ", err)
	fmt.Println("state:", ins.Current())
	fmt.Println("-------------------")

	err = ins.Send(context.Background(), example.FooEventFail{Err: errors.New("whoops")})
	fmt.Println("err:  ", err)
	fmt.Println("state:", ins.Current())
	fmt.Println("-------------------")
}
