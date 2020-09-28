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
	go globalQueue.Run(context.TODO())
}

// Enqueues a task in the global queue.
func Enqueue(t *Task) {
	ensureQueue()
	globalQueue.Enqueue(t)
}

func Submit(fn func(ctx context.Context) error) *Task {
	ensureQueue()
	return globalQueue.Submit(fn)
}
