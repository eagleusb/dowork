# dowork [![godoc](https://godoc.org/git.sr.ht/~sircmpwn/dowork?status.svg)](https://godoc.org/git.sr.ht/~sircmpwn/dowork) [![builds.sr.ht status](https://builds.sr.ht/~sircmpwn/dowork.svg)](https://builds.sr.ht/~sircmpwn/dowork)

dowork is a generic task queueing system for Go programs. It queues, executes,
and reschedules tasks in Goroutine in-process.

A global task queue is provided for simple use-cases. To use it:

```go
import (
  "context"

  "git.sr.ht/~sircmpwn/dowork"
)

// ...
work.Submit(func(ctx context.Context) error {
  // Thing which might fail...
  return nil
})
```

This task will be executed in the background and automatically retried with an
exponential backoff, up to a maximum number of attempts. The first time a task
is submitted to the global queue, it will be initialized and start running in
the background.

To customize options like maximum retries and timeouts, use work.Enqueue:

```go
task := work.NewTask(func(ctx context.Context) error {
  // ...
}).Retries(5).MaxTimeout(10 * time.Minute)
work.Enqueue(task)
```

You may also manage your own work queues. Use `NewQueue()` to obtain a queue,
`queue.Dispatch()` to execute all overdue tasks, and `queue.Run()` to spin up a
goroutine and start dispatching tasks automatically.

## Distributed task queues

No such functionality is provided OOTB, but you may find this package useful in
doing the actual queueing work for your own distributed work queue. Such an
integeration is left as an exercise to the reader.
