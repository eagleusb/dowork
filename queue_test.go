package work

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	now := time.Now().UTC()
	q := NewQueue()
	q.Now(func() time.Time {
		return now
	})

	var calls int
	q.Task(func() error {
		calls++
		return nil
	})

	var attempts int
	q.Task(func() error {
		calls++
		attempts++
		if attempts >= 2 {
			return nil
		}
		return errors.New("error")
	})

	q.Dispatch()
	assert.Equal(t, 2, calls)

	q.Dispatch()
	assert.Equal(t, 2, calls)

	now = now.Add(5 * time.Minute)

	q.Dispatch()
	assert.Equal(t, 3, calls)

	now = now.Add(5 * time.Minute)

	q.Dispatch()
	assert.Equal(t, 3, calls)
}

func TestTasksQueueingTasks(t *testing.T) {
	q := NewQueue()

	var calls int
	q.Task(func() error {
		// Should not deadlock
		q.Task(func() error {
			calls++
			return nil
		})
		calls++
		return nil
	})

	q.Dispatch()
	assert.Equal(t, 1, calls)

	q.Dispatch()
	assert.Equal(t, 2, calls)
}