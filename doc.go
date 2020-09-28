// dowork is a generic task queueing system for Go programs. It queues,
// executes, and reschedules tasks in Goroutine in-process.
//
// A global task queue is provided for simple use-cases. To use it:
//
//	import "git.sr.ht/~sircmpwn/dowork"
//
//	// ...
//	work.Submit(func() error {
//		// Thing which might fail...
//		return nil
//	})
//
// The first time a task is submitted to the global queue, it will be
// initialized and start running in the background.
//
// To customize the behavior, use work.Enqueue:
//
//	task := work.NewTask(func() error {
//		// ...
//	}).Retries(5).MaxTimeout(10 * time.Minute)
//	work.Enqueue(task)
//
// You may also manage your own work queues. Use NewQueue() to obtain a queue,
// queue.Dispatch() to execute all overdue tasks, and queue.Run() to spin up a
// Goroutine and start dispatching tasks automatically.
package work
