package k8s

import (
	"time"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Flow is a helper interface to interact with current reconcile loop and providing a
// logger method to get current logger (with some attached key-values).
type Flow interface {
	// Logger returns the logger
	Logger() logr.Logger

	// RetryAfter requeue the reconcile request after given duration and log the message and key-values.
	RetryAfter(duration time.Duration, msg string, kvs ...interface{}) (reconcile.Result, error)

	// Retry the reconcile request immediately and log the message and key-values.
	Retry(msg string, kvs ...interface{}) (reconcile.Result, error)

	// Continue the steps and log the message and key-values.
	Continue(msg string, kvs ...interface{}) (reconcile.Result, error)

	// Pass continues the steps without any logs.
	Pass() (reconcile.Result, error)

	// Wait breaks the current reconcile and wait for another reconcile request, and log the message and key-values.
	Wait(msg string, kvs ...interface{}) (reconcile.Result, error)

	// Break is same as Wait.
	Break(msg string, kvs ...interface{}) (reconcile.Result, error)

	// Error breaks the current reconcile with error and log the message and key-values.
	Error(err error, msg string, kvs ...interface{}) (reconcile.Result, error)

	// RetryErr is like Error but without returning error to the controller framework, but
	// give it a chance to retry later. Default retry period is 1s.
	RetryErr(err error, msg string, kvs ...interface{}) (reconcile.Result, error)

	// WithLogger return a flow binding to the old but with a new logger.
	WithLogger(log logr.Logger) Flow

	// WithLoggerValues returns a flow binding to the old but logs with extra values.
	WithLoggerValues(keyAndValues ...interface{}) Flow
}

type innerFlow interface {
	// BreakLoop indicates if we should return from current reconcile.
	BreakLoop() bool

	// SetLogger set the logger.
	SetLogger(logr.Logger)

	Flow
}

type flow struct {
	retryAfter time.Duration
	breakLoop  *bool
	logger     logr.Logger
}

func (f *flow) WithLogger(log logr.Logger) Flow {
	return &flow{
		breakLoop: f.breakLoop,
		logger:    log.WithCallDepth(1),
	}
}

func (f *flow) WithLoggerValues(keyAndValues ...interface{}) Flow {
	return &flow{
		breakLoop: f.breakLoop,
		logger:    f.logger.WithValues(keyAndValues...),
	}
}

func (f *flow) Logger() logr.Logger {
	return f.logger
}

func (f *flow) markBreak() {
	*f.breakLoop = true
}

func (f *flow) BreakLoop() bool {
	return *f.breakLoop
}

func (f *flow) RetryAfter(duration time.Duration, msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Info(msg, kvs...)
	return reconcile.Result{RequeueAfter: duration}, nil
}

func (f *flow) Retry(msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Info(msg, kvs...)
	return reconcile.Result{Requeue: true}, nil
}

func (f *flow) Continue(msg string, kvs ...interface{}) (reconcile.Result, error) {
	f.logger.Info(msg, kvs...)
	return reconcile.Result{}, nil
}

func (f *flow) Pass() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (f *flow) Wait(msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Info(msg, kvs...)
	return reconcile.Result{}, nil
}

func (f *flow) Break(msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Info(msg, kvs...)
	return reconcile.Result{}, nil
}

func (f *flow) Error(err error, msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Error(err, msg, kvs...)
	return reconcile.Result{}, err
}

func (f *flow) RetryErr(err error, msg string, kvs ...interface{}) (reconcile.Result, error) {
	defer f.markBreak()

	f.logger.Error(err, msg, kvs...)
	return reconcile.Result{RequeueAfter: f.retryAfter}, nil
}

func (f *flow) SetLogger(l logr.Logger) {
	f.logger = l
}

func newFlow(l logr.Logger) *flow {
	breakLoop := false
	return &flow{
		retryAfter: 1 * time.Second,
		breakLoop:  &breakLoop,
		logger:     l,
	}
}
