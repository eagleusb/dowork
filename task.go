package work

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	// Returned when a task is attempted which was already successfully completed.
	ErrAlreadyComplete = errors.New("This task was already successfully completed once")

	// If this is returned from a task function, the task shall not be re-attempted.
	ErrDoNotReattempt = errors.New("This task should not be re-attempted")

	// This task has been attempted too many times.
	ErrMaxRetriesExceeded = errors.New("The maximum retries for this task has been exceeded")

	// Set this function to influence the clock that will be used for
	// scheduling re-attempts.
	Now = func() time.Time {
		return time.Now().UTC()
	}
)

// Stores state for a task which shall be or has been executed. Each task may
// only be executed successfully once.
type Task struct {
	Metadata map[string]interface{}

	after       func(ctx context.Context, err error)
	attempts    int
	done        bool
	err         error
	fn          func(ctx context.Context) error
	nextAttempt time.Time

	maxAttempts int
	maxTimeout  time.Duration
	within      time.Duration
}

// Creates a new task for a given function.
func NewTask(fn func(ctx context.Context) error) *Task {
	return &Task{
		Metadata: make(map[string]interface{}),

		fn:          fn,
		maxAttempts: 10,
		maxTimeout:  30 * time.Minute,
	}
}

// Attempts to execute this task.
//
// If successful, the zero time and nil are returned.
//
// Otherwise, the error returned from the task function is returned to the
// caller. If an error is returned for which errors.Is(err, ErrDoNotReattempt)
// is true, the caller should not call Attempt again.
func (t *Task) Attempt(ctx context.Context) (time.Time, error) {
	if t.done {
		if t.err == nil {
			return time.Time{}, ErrAlreadyComplete
		}
		return time.Time{}, t.err
	}

	t.attempts += 1
	if t.attempts > t.maxAttempts {
		t.err = ErrMaxRetriesExceeded
		t.done = true
		if t.after != nil {
			t.after(ctx, t.err)
			t.after = nil
		}
		return time.Time{}, ErrMaxRetriesExceeded
	}

	if errors.Is(t.err, ErrDoNotReattempt) {
		return time.Time{}, t.err
	}

	if t.within != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.within)
		defer cancel()
	}

	t.err = t.fn(ctx)
	if t.err == nil {
		t.done = true
		if t.after != nil {
			t.after(ctx, t.err)
			t.after = nil
		}
		return time.Time{}, nil
	}

	next := time.Duration(int(math.Pow(2, float64(t.attempts)))) * time.Minute
	if next > t.maxTimeout {
		next = t.maxTimeout
	}
	t.nextAttempt = Now().Add(next)
	return t.nextAttempt, t.err
}

// Set the maximum number of retries on failure, or -1 to attempt indefinitely.
// By default, a task will be retried a maximum of 10 times.
func (t *Task) Retries(n int) *Task {
	if n < -1 {
		panic(errors.New("Invalid input to Task.Retries"))
	}
	t.maxAttempts = n
	return t
}

// Sets the maximum timeout between retries, or zero to exponentially increase
// the timeout indefinitely. Defaults to 30 minutes.
func (t *Task) MaxTimeout(d time.Duration) *Task {
	if d < 0 {
		panic(errors.New("Invalid timeout provided to Task.MaxTimeout"))
	}
	t.maxTimeout = d
	return t
}

// Returns the number of times this task has been attempted
func (t *Task) Attempts() int {
	return t.attempts
}

// Returns the time the next attempt is scheduled for, or the zero value if it
// has not been attempted before.
func (t *Task) NextAttempt() time.Time {
	return t.nextAttempt
}

// Returns true if this task was completed, successfully or not.
func (t *Task) Done() bool {
	return t.done
}

// Sets a function which will be executed once the task is completed,
// successfully or not. The final result (nil or an error) is passed to the
// callee.
func (t *Task) After(fn func(ctx context.Context, err error)) *Task {
	if t.after != nil {
		panic(errors.New("This task already has an 'After' function assigned"))
	}
	t.after = fn
	return t
}

// Specifies an upper limit for the duration of each attempt.
func (t *Task) Within(deadline time.Duration) {
	t.within = deadline
}
