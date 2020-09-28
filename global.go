package work

var globalQueue *Queue

func ensureQueue() {
	if globalQueue != nil {
		return
	}
	globalQueue = NewQueue()
	go globalQueue.Run()
}

// Enqueues a task in the global queue.
func Enqueue(t *Task) {
	ensureQueue()
	globalQueue.Enqueue(t)
}

func Submit(fn func() error) *Task {
	ensureQueue()
	return globalQueue.Submit(fn)
}
