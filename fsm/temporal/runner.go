package temporal

import (
	"context"
	"github.com/lovromazgon/fsm/fsm"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"time"
)

type Helper struct {
	Data struct {
		State string
	}
}

func (h *Helper) State() string {
	return h.Data.State
}

func RegisterFSMWorkflow(fsm fsm.FSM, wr worker.WorkflowRegistry, ar worker.ActivityRegistry) {
	wr.RegisterWorkflowWithOptions(func(ctx workflow.Context) error {
		return FooFSM(ctx)
	}, workflow.RegisterOptions{Name: "foofsm"})
	ar.RegisterActivityWithOptions(func(ctx context.Context, helper *Helper) (string, error) {
		return fsm.Observe(ctx, helper)
	}, activity.RegisterOptions{Name: "Observe"})
	for name, sf := range fsm.StateFunctions() {
		ar.RegisterActivityWithOptions(func(ctx context.Context, helper *Helper) error {
			return sf(ctx, helper)
		}, activity.RegisterOptions{Name: name})
	}
}

func FooFSM(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("FSM workflow started")

	h := &Helper{}
	// Setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func() (string, error) {
		return h.State(), nil
	})
	if err != nil {
		logger.Error("SetQueryHandler failed", "Error", err)
		return err
	}

	tick := workflow.GetSignalChannel(ctx, "tick")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 0; i < 10; i++ {
		var state string
		err = workflow.ExecuteActivity(ctx, "Observe", h).Get(ctx, &state)
		if err != nil {
			logger.Error("Observe failed.", "Error", err)
			return err
		}
		h.Data.State = state
		err = workflow.ExecuteActivity(ctx, state, h).Get(ctx, nil)
		if err != nil {
			logger.Error("State failed.", "Error", err)
			return err
		}

		_, _ = tick.ReceiveWithTimeout(ctx, time.Second, nil)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	logger.Info("FSM workflow completed.")
	return nil
}
