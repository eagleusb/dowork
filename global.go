package work

import (
	"context"
)

var globalQueue *Queue

func ensureQueue() {
	if globalQueue != nil {
		return
	}
	globalQueue = NewQueue()
	globalQueue.Start(context.TODO())
}

// Ensures that the global queue is started
func Start() {
	ensureQueue()
}

// Enqueues a task in the global queue.
func Enqueue(t *Task) {
	ensureQueue()
	globalQueue.Enqueue(t)
}

// See (*Queue).Submit
func Submit(fn func(ctx context.Context) error) (*Task, error) {
	ensureQueue()
	return globalQueue.Submit(fn)
}

// Stops accepting new tasks and blocks until all queued tasks are completed.
func Shutdown() {
	globalQueue.Shutdown()
}
