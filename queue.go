package work

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ErrQueueShuttingDown = errors.New("Queue is shutting down; new tasks are not being accepted")

	tasksQueued = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_enqueued_total",
		Help: "Total number of tasks submitted to a task queue",
	}, []string{"queue"})
	tasksRun = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_run_total",
		Help: "Total number of tasks run by a task queue",
	}, []string{"queue"})
	tasksCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_completed_total",
		Help: "Total number of tasks completed by a task queue",
	}, []string{"queue"})
	tasksSucceeded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_succeeded_total",
		Help: "Total number of tasks which completed in a success state",
	}, []string{"queue"})
	tasksFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "queue_tasks_failed_total",
		Help: "Total number of tasks which completed in a failed state",
	}, []string{"queue"})

	taskAttempts = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "queue_task_attempts",
		Help:    "Number of attempts required for each task to be completed",
		Buckets: []float64{1, 2, 3, 5, 8, 13, 21},
	}, []string{"queue"})
	taskDurations = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "queue_task_duration_sections",
		Help:    "Duration in seconds of each task attempt",
		Buckets: []float64{10, 30, 60, 90, 120, 300, 600, 900, 1800},
	}, []string{"queue"})
)

type Queue struct {
	mutex sync.Mutex
	name  string
	tasks []*Task
	next  time.Time
	now   func() time.Time
	wake  chan interface{}
	wg    sync.WaitGroup

	accept   int32
	shutdown chan interface{}
	started  bool // Only used to enforce constraints
}

// Creates a new task queue. The name of the task queue is used in Prometheus
// label names and must match [a-zA-Z0-9:_] (snake case is used by convention).
func NewQueue(name string) *Queue {
	return &Queue{
		name: name,
		now: func() time.Time {
			return time.Now().UTC()
		},
		accept: 1,
	}
}

// Sets the function the queue will use to obtain the current time.
func (q *Queue) Now(now func() time.Time) {
	q.now = now
}

// Enqueues a task.
//
// An error will only be returned if the queue has been shut down.
func (q *Queue) Enqueue(t *Task) error {
	if atomic.LoadInt32(&q.accept) == 0 {
		return ErrQueueShuttingDown
	}

	tasksQueued.WithLabelValues(q.name).Add(1)

	q.mutex.Lock()
	q.tasks = append(q.tasks, t)
	q.mutex.Unlock()
	if q.wake != nil {
		q.wake <- nil
	}
	return nil
}

// Creates and enqueues a new task, returning the new task. Note that the
// caller cannot customize settings on the task without creating a race
// condition; so attempting to will panic. See NewTask and (*Queue).Enqueue to
// create tasks with customized options.
//
// An error will only be returned if the queue has been shut down.
func (q *Queue) Submit(fn TaskFunc) (*Task, error) {
	t := NewTask(fn)
	t.immutable = true
	err := q.Enqueue(t)
	return t, err
}

// Attempts any tasks which are due and updates the task schedule. Returns true
// if there is more work to do.
func (q *Queue) Dispatch(ctx context.Context) bool {
	next := time.Unix(1<<63-62135596801, 999999999) // "max" time
	now := q.now()

	// In order to avoid deadlocking if a task queues another task, we make a
	// copy of the task list and release the mutex while executing them.
	q.mutex.Lock()
	tasks := make([]*Task, len(q.tasks))
	copy(tasks, q.tasks)
	q.mutex.Unlock()

	for _, task := range tasks {
		due := task.NextAttempt().Before(now)
		if due {
			if task.Attempts() == 0 {
				tasksRun.WithLabelValues(q.name).Add(1)
			}
			timer := prometheus.NewTimer(
				taskDurations.WithLabelValues(q.name))
			n, _ := task.Attempt(ctx)
			timer.ObserveDuration()
			if !task.Done() && n.Before(next) {
				next = n
			}
		}
	}

	q.mutex.Lock()
	newTasks := make([]*Task, 0, len(q.tasks))
	for _, task := range q.tasks {
		if !task.Done() {
			newTasks = append(newTasks, task)
		} else {
			if task.Result() == nil {
				tasksSucceeded.WithLabelValues(q.name).Add(1)
			} else {
				tasksFailed.WithLabelValues(q.name).Add(1)
			}
			tasksCompleted.WithLabelValues(q.name).Add(1)
			taskAttempts.WithLabelValues(q.name).Observe(float64(task.Attempts()))
		}
	}
	q.tasks = newTasks
	q.mutex.Unlock()

	q.next = next
	return len(newTasks) != 0
}

// Runs the task queue. Blocks until the context is cancelled.
func (q *Queue) Run(ctx context.Context) {
	if q.wake != nil {
		panic(errors.New("This queue is already running on another goroutine"))
	}

	q.wake = make(chan interface{})
	q.shutdown = make(chan interface{})

	for {
		more := q.Dispatch(ctx)
		if atomic.LoadInt32(&q.accept) == 0 && !more {
			return
		}

		select {
		case <-time.After(q.next.Sub(q.now())):
			break
		case <-ctx.Done():
			return
		case <-q.wake:
			break
		case <-q.shutdown:
			atomic.StoreInt32(&q.accept, 0)
			break
		}
	}
}

// Starts the task queue in the background. If you wish to use the warm
// shutdown feature, you must use Start, not Run.
func (q *Queue) Start(ctx context.Context) {
	q.started = true
	q.wg.Add(1)
	go func() {
		q.Run(ctx)
		q.wg.Done()
	}()
}

// Stops accepting new tasks and blocks until all already-queued tasks are
// complete. The queue must have been started with Start, not Run.
func (q *Queue) Shutdown() {
	if !q.started {
		panic(errors.New("Attempted warm shutdown on queue which was not run with queue.Start(ctx)"))
	}
	q.shutdown <- nil
	q.wg.Wait()
}
