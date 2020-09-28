package work

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Queue struct {
	mutex sync.Mutex
	tasks []*Task
	next  time.Time
	now   func() time.Time
	wake  chan interface{}
}

// Creates a new task queue.
func NewQueue() *Queue {
	return &Queue{
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Sets the function the queue will use to obtain the current time.
func (q *Queue) Now(now func() time.Time) {
	q.now = now
}

// Enqueues a task.
func (q *Queue) Enqueue(t *Task) {
	q.mutex.Lock()
	q.tasks = append(q.tasks, t)
	q.mutex.Unlock()
	if q.wake != nil {
		q.wake <- nil
	}
}

// Creates and enqueues a new task, returning the new task.
func (q *Queue) Submit(fn func(ctx context.Context) error) *Task {
	t := NewTask(fn)
	q.Enqueue(t)
	return t
}

// Attempts any tasks which are due and updates the task schedule. Returns true
// if there is more work to do.
func (q *Queue) Dispatch(ctx context.Context) bool {
	var next time.Time
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
			n, _ := task.Attempt(ctx)
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

	for {
		q.Dispatch(ctx)

		select {
		case <-time.After(q.next.Sub(q.now())):
			break
		case <-ctx.Done():
			return
		case <-q.wake:
			break
		}
	}
}
