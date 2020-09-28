// dowork is a generic task queueing system for Go programs. It queues,
// executes, and reschedules tasks in Goroutine in-process.
//
// A global task queue is provided for simple use-cases. To use it:
//
//	import (
//		"context"
//
//		"git.sr.ht/~sircmpwn/dowork"
//	)
//
//	// ...
//	work.Submit(func(ctx context.Context) error {
//		// Thing which might fail...
//		return nil
//	})
//
// This task will be executed in the background and automatically retried with
// an exponential backoff, up to a maximum number of attempts. The first time a
// task is submitted to the global queue, it will be initialized and start
// running in the background.
//
// To customize options like maximum retries and timeouts, use work.Enqueue:
//
//	task := work.NewTask(func(ctx context.Context) error {
//		// ...
//	}).Retries(5).MaxTimeout(10 * time.Minute)
//	work.Enqueue(task)
//
// You may also manage your own work queues. Use NewQueue() to obtain a queue,
// queue.Dispatch() to execute all overdue tasks, and queue.Run() to spin up a
// goroutine and start dispatching tasks automatically.
package work
