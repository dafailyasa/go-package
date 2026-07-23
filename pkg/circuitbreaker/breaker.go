package circuitbreaker

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/gobreaker/v2"
)

type Settings struct {
	Name        string
	MaxFailures uint32
	Interval    time.Duration
	Timeout     time.Duration
}

func DefaultSettings(name string) Settings {
	return Settings{
		Name:        name,
		MaxFailures: 5,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
	}
}

type Breaker struct {
	cb     *gobreaker.CircuitBreaker[struct{}]
	logger zerolog.Logger
}

func NewBreaker(settings Settings, logger zerolog.Logger) *Breaker {
	cb := gobreaker.NewCircuitBreaker[struct{}](gobreaker.Settings{
		Name:        settings.Name,
		Interval:    settings.Interval,
		Timeout:     settings.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= settings.MaxFailures
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Info().
				Str("breaker", name).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("circuit breaker state changed")
		},
	})

	return &Breaker{
		cb:     cb,
		logger: logger,
	}
}

func (b *Breaker) Execute(fn func() error) error {
	_, err := b.cb.Execute(func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

func (b *Breaker) State() gobreaker.State {
	return b.cb.State()
}

func (b *Breaker) Name() string {
	return b.cb.Name()
}
