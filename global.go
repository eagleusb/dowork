package work

var globalQueue *Queue

func ensureQueue() {
	if globalQueue != nil {
		return
	}
	globalQueue = NewQueue()
	go globalQueue.Run()
}

func Enqueue(t *Task) {
	ensureQueue()
	globalQueue.Enqueue(t)
}
