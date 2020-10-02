package work

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now time.Time

func init() {
	now = time.Now().UTC()
	Now = func() time.Time {
		return now
	}
}

func TestAttempt(t *testing.T) {
	taskerr := errors.New("First attempt")
	attempts := 0
	task := NewTask(func(ctx context.Context) error {
		attempts += 1
		if attempts < 2 {
			return taskerr
		}
		return nil
	})

	due, err := task.Attempt(context.TODO())
	assert.Equal(t, taskerr, err)
	assert.Equal(t, now.Add(2 * time.Minute), due)

	_, err = task.Attempt(context.TODO())
	assert.Nil(t, err)
	assert.True(t, task.Done())

	_, err = task.Attempt(context.TODO())
	assert.Equal(t, ErrAlreadyComplete, err)
}

func TestNoReattempt(t *testing.T) {
	attempts := 0
	task := NewTask(func(ctx context.Context) error {
		attempts += 1
		if attempts >= 2 {
			t.Error("Task called repeatedly despite requesting no re-attempt")
		}
		return fmt.Errorf("Do not reattempt this task; %w", ErrDoNotReattempt)
	})

	_, err := task.Attempt(context.TODO())
	assert.True(t, errors.Is(err, ErrDoNotReattempt))

	_, err = task.Attempt(context.TODO())
	assert.True(t, errors.Is(err, ErrDoNotReattempt))
}

func TestBackoff(t *testing.T) {
	task := NewTask(func(ctx context.Context) error {
		return errors.New("error")
	})

	due, _ := task.Attempt(context.TODO())
	assert.Equal(t, now.Add(2 * time.Minute), due)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(4 * time.Minute), due)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(8 * time.Minute), due)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(16 * time.Minute), due)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(30 * time.Minute), due)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(30 * time.Minute), due)

	task.MaxTimeout(10 * time.Minute)

	due, _ = task.Attempt(context.TODO())
	assert.Equal(t, now.Add(10 * time.Minute), due)
}

func TestMaxAttempts(t *testing.T) {
	taskerr := errors.New("error")
	task := NewTask(func(ctx context.Context) error {
		return taskerr
	}).Retries(3)

	_, err := task.Attempt(context.TODO())
	assert.Equal(t, err, taskerr)

	_, err = task.Attempt(context.TODO())
	assert.Equal(t, err, taskerr)

	_, err = task.Attempt(context.TODO())
	assert.Equal(t, err, taskerr)

	_, err = task.Attempt(context.TODO())
	assert.Equal(t, err, ErrMaxRetriesExceeded)
}

func TestAfter(t *testing.T) {
	var (
		afterCalls int
		taskCalls  int
	)
	task := NewTask(func(ctx context.Context) error {
		taskCalls++
		if taskCalls >= 2 {
			return nil
		}
		return errors.New("error")
	}).After(func(ctx context.Context, err error) {
		afterCalls += 1
	})

	_, err := task.Attempt(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, 1, taskCalls)
	assert.Equal(t, 0, afterCalls)

	_, err = task.Attempt(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 2, taskCalls)
	assert.Equal(t, 1, afterCalls)

	task.Attempt(context.TODO())
	assert.Equal(t, 2, taskCalls)
	assert.Equal(t, 1, afterCalls)
}
