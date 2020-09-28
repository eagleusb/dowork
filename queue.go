package work

import (
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

// Enqueues a task and returns the queue.
func (q *Queue) Enqueue(t *Task) *Queue {
	q.mutex.Lock()
	q.tasks = append(q.tasks, t)
	q.mutex.Unlock()
	if q.wake != nil {
		q.wake <- nil
	}
	return q
}

// Creates and enqueues a new task, returning the new task.
func (q *Queue) Task(fn func() error) *Task {
	t := NewTask(fn)
	q.Enqueue(t)
	return t
}

// Attempts any tasks which are due and updates the task schedule.
func (q *Queue) Dispatch() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var next time.Time
	now := q.now()
	newTasks := make([]*Task, 0, len(q.tasks))
	for _, task := range q.tasks {
		due := task.NextAttempt().Before(now)
		if due {
			n, err := task.Attempt()
			if err == nil ||
				err == ErrDoNotReattempt ||
				err == ErrMaxRetriesExceeded {
				continue
			}
			if n.Before(next) {
				next = n
			}
		}
		newTasks = append(newTasks, task) // Retry
	}

	q.tasks = newTasks
	q.next = next
}

// Runs the task queue. Never returns.
func (q *Queue) Run() {
	if q.wake != nil {
		panic(errors.New("This queue is already running on another goroutine"))
	}

	q.wake = make(chan interface{})
	for {
		q.Dispatch()

		select {
		case <-time.After(q.next.Sub(q.now())):
			break
		case <-q.wake:
			break
		}
	}
}
