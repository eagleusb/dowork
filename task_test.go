package work

import (
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
	task := NewTask(func() error {
		attempts += 1
		if attempts < 2 {
			return taskerr
		}
		return nil
	})

	due, err := task.Attempt()
	assert.Equal(t, taskerr, err)
	assert.Equal(t, now.Add(2 * time.Minute), due)

	_, err = task.Attempt()
	assert.Nil(t, err)

	_, err = task.Attempt()
	assert.Equal(t, ErrAlreadyComplete, err)
}

func TestNoReattempt(t *testing.T) {
	attempts := 0
	task := NewTask(func() error {
		attempts += 1
		if attempts >= 2 {
			t.Error("Task called repeatedly despite requesting no re-attempt")
		}
		return fmt.Errorf("Do not reattempt this task; %w", ErrDoNotReattempt)
	})

	_, err := task.Attempt()
	assert.True(t, errors.Is(err, ErrDoNotReattempt))

	_, err = task.Attempt()
	assert.True(t, errors.Is(err, ErrDoNotReattempt))
}

func TestBackoff(t *testing.T) {
	task := NewTask(func() error {
		return errors.New("error")
	})

	due, _ := task.Attempt()
	assert.Equal(t, now.Add(2 * time.Minute), due)

	due, _ = task.Attempt()
	assert.Equal(t, now.Add(4 * time.Minute), due)

	due, _ = task.Attempt()
	assert.Equal(t, now.Add(8 * time.Minute), due)

	due, _ = task.Attempt()
	assert.Equal(t, now.Add(16 * time.Minute), due)

	due, _ = task.Attempt()
	assert.Equal(t, now.Add(30 * time.Minute), due)

	due, _ = task.Attempt()
	assert.Equal(t, now.Add(30 * time.Minute), due)
}
