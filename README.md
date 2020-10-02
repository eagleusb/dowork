# dowork [![godoc](https://godoc.org/git.sr.ht/~sircmpwn/dowork?status.svg)](https://godoc.org/git.sr.ht/~sircmpwn/dowork) [![builds.sr.ht status](https://builds.sr.ht/~sircmpwn/dowork.svg)](https://builds.sr.ht/~sircmpwn/dowork)

dowork is a generic task queueing system for Go programs. It queues, executes,
and reschedules tasks in Goroutine in-process.

A global task queue is provided for simple use-cases. To use it:

```go
import (
  "context"

  "git.sr.ht/~sircmpwn/dowork"
)

work.Submit(func(ctx context.Context) error {
    // ...do work...
    return nil
})
```

This task will be executed in the background. The first time a task is submitted
to the global queue, it will be initialized and start running in the background.

To customize options like maximum retries and timeouts, use work.Enqueue:

```go
task := work.NewTask(func(ctx context.Context) error {
    // ...
}).Retries(5).MaxTimeout(10 * time.Minute)
work.Enqueue(task)
task := work.NewTask(func(ctx context.Context) error {
    // ...
}).
    Retries(5).                   // Maximum number of attempts
    MaxTimeout(10 * time.Minute). // Maximum timeout between attempts
    Within(10 * time.Second).     // Deadline for each attempt
    After(func(ctx context.Context, err error) {
        // Executed once the task completes, successful or not
    })
work.Enqueue(task)
```

Retries are conducted with an exponential backoff.

You may also manage your own work queues. Use `NewQueue()` to obtain a queue,
`(*Queue).Dispatch()` to execute all overdue tasks, and `(*Queue).Start()` to
spin up a goroutine and start dispatching tasks automatically.

Use `work.Shutdown()` or `(*Queue).Shutdown()` to perform a soft shutdown of the
queue, which will stop accepting new tasks and block until all already-queued
tasks complete.

## Distributed task queues

No such functionality is provided OOTB, but you may find this package useful in
doing the actual queueing work for your own distributed work queue. Such an
integeration is left as an exercise to the reader.

## Instrumentation

Instrumentation is provided via [Prometheus][prom] and the
[client_golang][client_golang] library. Relevant metrics are prefixed with
`queue_` in the metric name.

[prom]: https://prometheus.io
[client_golang]: https://github.com/prometheus/client_golang

A common pattern for soft restarting web servers is to shut down the
[http.Server][http.Server], allowing the new web server process to start up and
begin accepting new connections, then allow the queue to finish executing any
pending tasks before terminating the process. If this describes your program,
note that you may want to provide Prometheus metrics on a secondary http.Server
on a random port, so that you may monitor the queue shutdown. Something similar
to the following will set up a secondary HTTP server for this purpose:

[http.Server]: https://golang.org/pkg/net/http/#Server

```go
import (
    "log"
    "net"
    "net/http"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

mux := &http.ServeMux{}
mux.Handle("/metrics", promhttp.Handler())
server := &http.Server{Handler: mux}
listen, err := net.Listen("tcp", ":0")
if err != nil {
    panic(err)
}
log.Printf("Prometheus listening on :%d", listen.Addr().(*net.TCPAddr).Port)
go server.Serve(listen)
```
