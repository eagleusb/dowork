package work

import (
	"errors"
	"math"
	"time"
)

var (
	// Returned when a task is attempted which was already successfully completed
	ErrAlreadyComplete = errors.New("This task was already successfully completed once")
	ErrDoNotReattempt = errors.New("This task should not be re-attempted")

	Now = func() time.Time {
		return time.Now().UTC()
	}
)

// Stores state for a task which shall be or has been executed. Each task may
// only be executed successfully once.
type Task struct {
	fn func() error

	attempts    int
	err         error
	metadata    map[string]string
	nextAttempt time.Time
}

func NewTask(fn func() error) *Task {
	return &Task{fn: fn}
}

// Attempts to execute this task.
//
// If successful, the zero time and nil are returned.
//
// Otherwise, the error returned from the task function is returned to the
// caller. If an error is returned for which errors.Is(err, ErrDoNotReattempt)
// is true, the caller should not call Attempt again.
func (t *Task) Attempt() (time.Time, error) {
	if t.err == nil && t.attempts > 0 {
		return time.Time{}, ErrAlreadyComplete
	}
	if errors.Is(t.err, ErrDoNotReattempt) {
		return time.Time{}, t.err
	}
	t.attempts += 1
	t.err = t.fn()
	if t.err == nil {
		return time.Time{}, nil
	}
	next := time.Duration(int(math.Pow(2, float64(t.attempts)))) * time.Minute
	if next > 30 * time.Minute {
		next = 30 * time.Minute
	}
	t.nextAttempt = Now().Add(next)
	return t.nextAttempt, t.err
}

// Returns the number of times this task has been attempted
func (t *Task) Attempts() int {
	return t.attempts
}

// Returns the result of the last attempt, or nil if never attempted.
func (t *Task) Result() error {
	return t.err
}
