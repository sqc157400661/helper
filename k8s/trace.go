package k8s

import "github.com/google/uuid"

type tracer struct {
	stepIndex int
	id        string
}

func (t *tracer) currentStepIndex() int {
	return t.stepIndex
}

func (t *tracer) markStepDone() {
	t.stepIndex++
}

func newTracer() *tracer {
	return &tracer{
		stepIndex: 0,
		id:        uuid.New().String(),
	}
}
