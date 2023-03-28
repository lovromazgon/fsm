package example

import (
	"context"
	"fmt"
	"github.com/lovromazgon/fsm/fsm"
	"testing"
)

type Helper struct {
	// Data holds all data and prevents collisions with interface methods.
	Data struct {
		State string
	}
}

func (h *Helper) State() string {
	return h.Data.State
}

func TestLint(t *testing.T) {
	fsm.Lint(t, &FooFSM{})
}

func TestFoo(t *testing.T) {
	ctx := context.Background()
	var f fsm.FSM = &FooFSM{}
	sfs := f.StateFunctions()

	h := &Helper{}

	for i := 0; i < 10; i++ {
		s, err := f.Observe(ctx, h)
		if err != nil {
			t.Fatal(err)
		}
		if s == h.Data.State {
			fmt.Println("same state as before")
		} else {
			fmt.Printf("new state: %v\n", s)
			h.Data.State = s
		}
		err = sfs[h.Data.State](ctx, h)
		if err != nil {
			t.Fatal(err)
		}
	}
}
