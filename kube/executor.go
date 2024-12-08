package kube

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Executor interface {
	Execute(rc ReconcileContext, task *Task) (reconcile.Result, error)
}

type executor struct {
	logger logr.Logger
	flow   innerFlow
	tracer *tracer
	debug  bool
}

func (e *executor) isDebugEnabled() bool {
	return e.debug
}

func (e *executor) handlePanic(r interface{}) error {
	var err error

	// Extract error from panic
	switch r.(type) {
	case error:
		err = r.(error)
	case string:
		err = fmt.Errorf(r.(string))
	default:
		err = fmt.Errorf("%+v", r)
	}
	e.logger.Error(err, "Panic detected, recovered and return error")

	return err
}

func (e *executor) prepare(rc ReconcileContext, step Step, log logr.Logger) {
	name := step.Name()

	log = log.WithValues("action", name, "step", e.tracer.currentStepIndex())
	e.flow.SetLogger(log)

	if e.isDebugEnabled() || rc.Debug() {
		log.WithName("trace").Info("BEGIN")
	}
}

func (e *executor) done(rc ReconcileContext, err error, deferred bool, last bool) {
	e.tracer.markStepDone()

	if e.isDebugEnabled() || rc.Debug() {
		log := e.flow.Logger().WithName("trace")
		if err != nil {
			log.Info("ERROR", "err", err.Error())
		} else if last {
			log.Info("COMPLETE")
		} else if deferred {
			log.Info("CONTINUE [DEFER]")
		} else if e.flow.BreakLoop() {
			log.Info("BREAK")
		} else {
			log.Info("CONTINUE")
		}
	}
}

func (e *executor) execute(rc ReconcileContext, step Step, log logr.Logger, deferred bool, last bool) (result reconcile.Result, err error) {
	e.prepare(rc, step, log)

	defer e.done(rc, err, deferred, last)

	return step.Execute(rc, e.flow)
}

func (e *executor) executeDeferredSteps(rc ReconcileContext, task *Task) error {
	log := e.logger.WithValues("defer_exec", true)

	errs := make([]string, 0)
	for task.hasNextDeferredStep() {
		step := task.nextDeferredStep()
		_, err := e.execute(rc, step, log, true, !task.hasNextDeferredStep())

		// Never breaks the reconciliation flow, each deferred step will be executed.
		if err != nil {
			// e.logger.Error(err, "Err detected in deferred action.")
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New("Err in deferred actions: " + strings.Join(errs, ", "))
	}
	return nil
}

func (e *executor) Execute(rc ReconcileContext, task *Task) (result reconcile.Result, err error) {
	// Handle panic.
	defer func() {
		if r := recover(); r != nil {
			err = e.handlePanic(r)
		}
	}()

	// Handle force requeue after.
	defer func() {
		forceRequeueAfter := rc.ForceRequeueAfter()
		if forceRequeueAfter > 0 {
			if err != nil {
				// Reset error (should have been logged by flow.Error()).
				err = nil
				result = reconcile.Result{RequeueAfter: forceRequeueAfter}
			} else {
				// Set the requeue after if not set or larger than the task's.
				if result.RequeueAfter == 0 || result.RequeueAfter > forceRequeueAfter {
					result.RequeueAfter = forceRequeueAfter
				}
			}
		}
	}()

	// Handle deferred actions.
	defer func() {
		err1 := e.executeDeferredSteps(rc, task)
		if err1 != nil {
			err = err1
		}
	}()

	// Execute steps.
	for task.hasNextStep() {
		step := task.nextStep()
		result, err = e.execute(rc, step, e.logger, false, !task.hasNextStep() && !task.hasNextDeferredStep())
		if e.flow.BreakLoop() {
			return
		}
	}

	return
}

type ExecutorOption func(e *executor)

func Debug(e *executor) {
	e.debug = true
}

func NewExecutor(logger logr.Logger, opts ...ExecutorOption) Executor {
	tracer := newTracer()

	exec := &executor{
		logger: logger.WithValues("trace", tracer.id),
		tracer: tracer,
	}

	exec.flow = newFlow(exec.logger)

	for _, opt := range opts {
		opt(exec)
	}

	return exec
}
