package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/lovromazgon/fsm"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func RunWorker[S fsm.State, O any, FSM fsm.FSM[S, O]](def FSM, opt ...client.Options) error {
	// The client and worker are heavyweight objects that should be created once per process.
	var co client.Options
	if len(opt) > 0 {
		co = opt[0]
	}

	c, err := client.Dial(co)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	w := worker.New(c, "fsm", worker.Options{})
	RegisterFSMWorkflow[S, O, FSM](def, w, w)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	return nil
}

func RegisterFSMWorkflow[S fsm.State, O any, FSM fsm.FSM[S, O]](def FSM, wr worker.WorkflowRegistry, ar worker.ActivityRegistry) {
	transitions := def.Transitions()
	wr.RegisterWorkflowWithOptions(func(ctx workflow.Context, ins FSM) error {
		return newWorkflowInstance[S, O, FSM](ins).Run(ctx)
	}, workflow.RegisterOptions{Name: workflowNameForFSM(def)})
	ar.RegisterActivityWithOptions(func(ctx context.Context, ins FSM, h *Helper[S]) (observeDTO[S, O, FSM], error) {
		o, err := ins.Observe(ctx, h)
		return observeDTO[S, O, FSM]{
			Observation: o,
			Instance:    ins,
		}, err
	}, activity.RegisterOptions{Name: "Observe"})
	ar.RegisterActivityWithOptions(func(ctx context.Context, ins FSM, h *Helper[S], o O) (transitionDTO[S, O, FSM], error) {
		for _, t := range transitions {
			if t.From == h.State && t.Condition(o) {
				// this triggers transition
				err := ins.Transition(ctx, h, t, o)
				return transitionDTO[S, O, FSM]{
					State:    t.To,
					Instance: ins,
				}, err
			}
		}
		return transitionDTO[S, O, FSM]{
			State:    h.Current(),
			Instance: ins,
		}, nil // no transition
	}, activity.RegisterOptions{Name: "Transition"})
	ar.RegisterActivityWithOptions(func(ctx context.Context, ins FSM, h *Helper[S], o O) (actionDTO[S, O, FSM], error) {
		err := ins.Action(ctx, h, o)
		return actionDTO[S, O, FSM]{
			Instance: ins,
		}, err
	}, activity.RegisterOptions{Name: "Action"})
}

type workflowInstance[S fsm.State, O any, FSM fsm.FSM[S, O]] struct {
	logger log.Logger
	tick   workflow.ReceiveChannel

	instance FSM
	helper   *Helper[S]
}

func newWorkflowInstance[S fsm.State, O any, FSM fsm.FSM[S, O]](ins FSM) *workflowInstance[S, O, FSM] {
	return &workflowInstance[S, O, FSM]{
		instance: ins,
		helper: &Helper[S]{
			State: ins.States()[0], // initial state
		},
	}
}

func (w *workflowInstance[S, O, FSM]) init(ctx workflow.Context) (workflow.Context, error) {
	w.logger = workflow.GetLogger(ctx)
	w.logger.Info("Initializing FSM workflow ...")

	// Setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func() (S, error) {
		return w.helper.Current(), nil
	})
	if err != nil {
		w.logger.Error("SetQueryHandler failed", "Error", err)
		return nil, err
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	w.tick = workflow.GetSignalChannel(ctx, "tick")

	w.logger.Info("FSM workflow initialized")
	return ctx, nil
}

func (w *workflowInstance[S, O, FSM]) Run(ctx workflow.Context) error {
	ctx, err := w.init(ctx)
	if err != nil {
		return err
	}

	w.logger.Info("FSM workflow started")
	for {
		o, err := w.observe(ctx)
		if err != nil {
			w.logger.Error("Observe failed.", "Error", err)
			return err
		}

		s, err := w.transition(ctx, o)
		if err != nil {
			w.logger.Error("Transition failed.", "Error", err)
			return err
		}

		w.helper.State = s
		err = w.action(ctx, o)
		if err != nil {
			w.logger.Error("Action failed.", "Error", err)
			return err
		}

		if w.helper.State.Done() {
			break
		}

		_, _ = w.tick.ReceiveWithTimeout(ctx, time.Second, nil)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if w.helper.State.Failed() {
		w.logger.Warn("FSM workflow failed.")
		return fmt.Errorf("workflow failed with status %q", w.helper.State)
	}

	w.logger.Info("FSM workflow completed.")
	return nil
}

func (w *workflowInstance[S, O, FSM]) observe(ctx workflow.Context) (O, error) {
	var out observeDTO[S, O, FSM]
	err := workflow.ExecuteActivity(ctx, "Observe", w.instance, w.helper).Get(ctx, &out)
	if err != nil {
		return empty[O](), err
	}
	w.instance = out.Instance
	return out.Observation, nil
}

func (w *workflowInstance[S, O, FSM]) transition(ctx workflow.Context, o O) (S, error) {
	var out transitionDTO[S, O, FSM]
	err := workflow.ExecuteActivity(ctx, "Transition", w.instance, w.helper, o).Get(ctx, &out)
	if err != nil {
		return empty[S](), err
	}
	w.instance = out.Instance
	return out.State, nil
}

func (w *workflowInstance[S, O, FSM]) action(ctx workflow.Context, o O) error {
	var out actionDTO[S, O, FSM]
	err := workflow.ExecuteActivity(ctx, "Action", w.instance, w.helper, o).Get(ctx, &out)
	if err != nil {
		return err
	}
	w.instance = out.Instance
	return nil
}

type Helper[S fsm.State] struct {
	State S
}

func (f *Helper[S]) Current() S {
	return f.State
}

type observeDTO[S fsm.State, O any, FSM fsm.FSM[S, O]] struct {
	Observation O
	Instance    FSM
}

type transitionDTO[S fsm.State, O any, FSM fsm.FSM[S, O]] struct {
	State    S
	Instance FSM
}

type actionDTO[S fsm.State, O any, FSM fsm.FSM[S, O]] struct {
	Instance FSM
}

func empty[T any]() T { var t T; return t }
