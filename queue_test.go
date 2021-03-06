package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	now := time.Now().UTC()
	q := NewQueue("test")
	q.Now(func() time.Time {
		return now
	})

	var calls int
	q.Submit(func(ctx context.Context) error {
		calls++
		return nil
	})

	var attempts int
	task := NewTask(func(ctx context.Context) error {
		calls++
		attempts++
		if attempts >= 2 {
			return nil
		}
		return errors.New("error")
	}).Retries(10)
	q.Enqueue(task)

	q.Dispatch(context.TODO())
	assert.Equal(t, 2, calls)

	q.Dispatch(context.TODO())
	assert.Equal(t, 2, calls)

	now = now.Add(5 * time.Minute)

	q.Dispatch(context.TODO())
	assert.Equal(t, 3, calls)

	now = now.Add(5 * time.Minute)

	q.Dispatch(context.TODO())
	assert.Equal(t, 3, calls)
}

func TestTasksQueueingTasks(t *testing.T) {
	q := NewQueue("test")

	var calls int
	q.Submit(func(ctx context.Context) error {
		// Should not deadlock
		q.Submit(func(ctx context.Context) error {
			calls++
			return nil
		})
		calls++
		return nil
	})

	q.Dispatch(context.TODO())
	assert.Equal(t, 1, calls)

	q.Dispatch(context.TODO())
	assert.Equal(t, 2, calls)
}

func TestRun(t *testing.T) {
	q := NewQueue("test")
	var (
		calledA bool
		calledB bool
	)
	q.Submit(func(ctx context.Context) error {
		calledA = true
		return errors.New("error")
	})
	ctx, cancel := context.WithCancel(context.TODO())
	go q.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	assert.True(t, calledA)

	q.Submit(func(ctx context.Context) error {
		calledB = true
		return errors.New("error")
	})
	time.Sleep(50 * time.Millisecond)
	assert.True(t, calledB)
	cancel()
}

func TestStart(t *testing.T) {
	q := NewQueue("test")
	var (
		calledA bool
		calledB bool
	)
	q.Submit(func(ctx context.Context) error {
		calledA = true
		return nil
	})
	q.Start(context.TODO())
	time.Sleep(50 * time.Millisecond)
	assert.True(t, calledA)

	q.Submit(func(ctx context.Context) error {
		calledB = true
		return nil
	})
	q.Shutdown()
	time.Sleep(50 * time.Millisecond)
	assert.True(t, calledB)

	_, err := q.Submit(func(ctx context.Context) error {
		t.Error("Should not be called")
		return nil
	})
	assert.Equal(t, ErrQueueShuttingDown, err)
}
