package temporal

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/lovromazgon/fsm"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type Helper[S fsm.State] struct {
	State S
}

func (f *Helper[S]) Current() S {
	return f.State
}

func RunWorker[S fsm.State, O any, I fsm.Instance[S, O]](def fsm.Definition[S, O, I], opt ...client.Options) {
	// The client and worker are heavyweight objects that should be created once per process.
	var co client.Options
	if len(opt) > 0 {
		co = opt[0]
	}

	c, err := client.Dial(co)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "fsm", worker.Options{})
	RegisterFSMWorkflow(def, w, w)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func RegisterFSMWorkflow[S fsm.State, O any, I fsm.Instance[S, O]](def fsm.Definition[S, O, I], wr worker.WorkflowRegistry, ar worker.ActivityRegistry) {
	transitions := def.Transitions()
	states := def.States()
	wr.RegisterWorkflowWithOptions(func(ctx workflow.Context) error {
		return run[S, O, I](ctx, def.New(), states[0])
	}, workflow.RegisterOptions{Name: workflowNameForFSM(def)})
	ar.RegisterActivityWithOptions(func(ctx context.Context, ins I, h *Helper[S]) (activityDTO[S, O, I], error) {
		o, err := ins.Observe(ctx, h)
		return activityDTO[S, O, I]{
			Observation: o,
			Instance:    ins,
		}, err
	}, activity.RegisterOptions{Name: "Observe"})
	ar.RegisterActivityWithOptions(func(ctx context.Context, ins I, h *Helper[S], o O) (activityDTO[S, O, I], error) {
		for _, t := range transitions {
			if t.From == h.State && t.Condition(o) {
				// this triggers transition
				err := ins.Transition(ctx, h, t, o)
				return activityDTO[S, O, I]{
					State:    t.To,
					Instance: ins,
				}, err
			}
		}
		return activityDTO[S, O, I]{
			State:    h.Current(),
			Instance: ins,
		}, nil // no transition
	}, activity.RegisterOptions{Name: "Transition"})
	for _, s := range states {
		ar.RegisterActivityWithOptions(func(ctx context.Context, ins I, h *Helper[S], o O) (activityDTO[S, O, I], error) {
			err := ins.Action(ctx, h, o)
			return activityDTO[S, O, I]{
				Instance: ins,
			}, err
		}, activity.RegisterOptions{Name: activityNameForState(s)})
	}
}

func run[S fsm.State, O any, I fsm.Instance[S, O]](ctx workflow.Context, ins I, initialState S) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("FSM workflow started")

	h := &Helper[S]{
		State: initialState,
	}
	// Setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func() (S, error) {
		return h.Current(), nil
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

	for {
		var out activityDTO[S, O, I]
		err = workflow.ExecuteActivity(ctx, "Observe", ins, h).Get(ctx, &out)
		if err != nil {
			logger.Error("Observe failed.", "Error", err)
			return err
		}

		o := out.Observation
		ins = out.Instance
		out = activityDTO[S, O, I]{}

		err = workflow.ExecuteActivity(ctx, "Transition", ins, h, o).Get(ctx, &out)
		if err != nil {
			logger.Error("Transition failed.", "Error", err)
			return err
		}

		ins = out.Instance
		h.State = out.State
		out = activityDTO[S, O, I]{}

		stateActivity := activityNameForState(h.State)

		err = workflow.ExecuteActivity(ctx, stateActivity, ins, h, o).Get(ctx, &out)
		if err != nil {
			logger.Error("State action failed.", "Action", stateActivity, "Error", err)
			return err
		}

		ins = out.Instance

		if h.State.Done() {
			break
		}

		_, _ = tick.ReceiveWithTimeout(ctx, time.Second, nil)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if h.State.Failed() {
		logger.Warn("FSM workflow failed.")
		return fmt.Errorf("workflow failed with status %q", h.State)
	}

	logger.Info("FSM workflow completed.")
	return nil
}

func activityNameForState[S ~string](s S) string {
	return string(s) + "::action"
}

func workflowNameForFSM[S fsm.State, O any, I fsm.Instance[S, O]](def fsm.Definition[S, O, I]) string {
	return reflect.TypeOf(def).String()
}

type activityDTO[S fsm.State, O any, I fsm.Instance[S, O]] struct {
	State       S
	Observation O
	Instance    I
}
