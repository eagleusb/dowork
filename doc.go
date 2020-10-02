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
//	work.Submit(func(ctx context.Context) error {
//		// ...do work...
//		return nil
//	})
//
// This task will be executed in the background. The first time a task is
// submitted to the global queue, it will be initialized and start running in
// the background.
//
// To customize options like maximum retries and timeouts, use work.Enqueue:
//
//	task := work.NewTask(func(ctx context.Context) error {
//		// ...
//	}).
//		Retries(5).			// Maximum number of attempts
//		MaxTimeout(10 * time.Minute).	// Maximum timeout between attempts
//		Within(10 * time.Second).	// Deadline for each attempt
//		After(func(ctx context.Context, err error) {
//			// Executed once the task completes, successful or not
//		})
//	work.Enqueue(task)
//
// Retries are conducted with an exponential backoff.
//
// You may also manage your own work queues. Use NewQueue() to obtain a queue,
// (*Queue).Dispatch() to execute all overdue tasks, and (*Queue).Start() to
// spin up a goroutine and start dispatching tasks automatically.
//
// Use work.Shutdown() or (*Queue).Shutdown() to perform a soft shutdown of the
// queue, which will stop accepting new tasks and block until all
// already-queued tasks complete.
package work
