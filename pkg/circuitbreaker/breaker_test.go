package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreaker_Execute_Success(t *testing.T) {
	logger := zerolog.Nop()
	breaker := NewBreaker(DefaultSettings("test"), logger)

	err := breaker.Execute(func() error {
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, gobreaker.StateClosed, breaker.State())
}

func TestBreaker_Execute_Error(t *testing.T) {
	logger := zerolog.Nop()
	breaker := NewBreaker(DefaultSettings("test"), logger)

	testErr := errors.New("operation failed")
	err := breaker.Execute(func() error {
		return testErr
	})
	assert.ErrorIs(t, err, testErr)
}

func TestBreaker_OpensAfterMaxFailures(t *testing.T) {
	settings := Settings{
		Name:        "test-open",
		MaxFailures: 3,
		Interval:    time.Minute,
		Timeout:     time.Minute,
	}
	logger := zerolog.Nop()
	breaker := NewBreaker(settings, logger)

	for i := 0; i < 3; i++ {
		_ = breaker.Execute(func() error {
			return errors.New("fail")
		})
	}

	assert.Equal(t, gobreaker.StateOpen, breaker.State())

	err := breaker.Execute(func() error {
		return nil
	})
	assert.Error(t, err, "should reject when open")
}

func TestBreaker_RecoveryAfterTimeout(t *testing.T) {
	settings := Settings{
		Name:        "test-recovery",
		MaxFailures: 1,
		Interval:    time.Minute,
		Timeout:     50 * time.Millisecond,
	}
	logger := zerolog.Nop()
	breaker := NewBreaker(settings, logger)

	_ = breaker.Execute(func() error {
		return errors.New("fail")
	})
	assert.Equal(t, gobreaker.StateOpen, breaker.State())

	time.Sleep(60 * time.Millisecond)

	err := breaker.Execute(func() error {
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, gobreaker.StateClosed, breaker.State())
}

func TestBreaker_RemainsOpenOnFailureAfterTimeout(t *testing.T) {
	settings := Settings{
		Name:        "test-remain-open",
		MaxFailures: 1,
		Interval:    time.Minute,
		Timeout:     50 * time.Millisecond,
	}
	logger := zerolog.Nop()
	breaker := NewBreaker(settings, logger)

	_ = breaker.Execute(func() error {
		return errors.New("fail")
	})
	assert.Equal(t, gobreaker.StateOpen, breaker.State())

	time.Sleep(60 * time.Millisecond)

	_ = breaker.Execute(func() error {
		return errors.New("fail again")
	})
	assert.Equal(t, gobreaker.StateOpen, breaker.State())
}

func TestBreaker_Name(t *testing.T) {
	logger := zerolog.Nop()
	breaker := NewBreaker(DefaultSettings("my-breaker"), logger)
	assert.Equal(t, "my-breaker", breaker.Name())
}

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings("test")
	assert.Equal(t, "test", s.Name)
	assert.Equal(t, uint32(5), s.MaxFailures)
	assert.Equal(t, 30*time.Second, s.Interval)
	assert.Equal(t, 60*time.Second, s.Timeout)
}

func TestBreaker_StateTransitions(t *testing.T) {
	settings := Settings{
		Name:        "test-transitions",
		MaxFailures: 2,
		Interval:    time.Minute,
		Timeout:     50 * time.Millisecond,
	}
	logger := zerolog.Nop()
	breaker := NewBreaker(settings, logger)

	assert.Equal(t, gobreaker.StateClosed, breaker.State())

	_ = breaker.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, gobreaker.StateClosed, breaker.State())

	_ = breaker.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, gobreaker.StateOpen, breaker.State())

	time.Sleep(60 * time.Millisecond)

	_ = breaker.Execute(func() error { return nil })
	assert.Equal(t, gobreaker.StateClosed, breaker.State())
}
