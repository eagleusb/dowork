package work

import (
	"errors"
	"math"
	"time"
)

var (
	// Returned when a task is attempted which was already successfully completed.
	ErrAlreadyComplete = errors.New("This task was already successfully completed once")

	// If this is returned from a task function, the task shall not be re-attempted.
	ErrDoNotReattempt = errors.New("This task should not be re-attempted")

	// Set this function to influence the clock that will be used for
	// scheduling re-attempts.
	Now = func() time.Time {
		return time.Now().UTC()
	}
)

// Stores state for a task which shall be or has been executed. Each task may
// only be executed successfully once.
type Task struct {
	Metadata map[string]string

	attempts    int
	err         error
	fn          func() error
	nextAttempt time.Time
}

// Creates a new task for a given function.
func NewTask(fn func() error) *Task {
	return &Task{
		Metadata: make(map[string]string),

		fn: fn,
	}
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
