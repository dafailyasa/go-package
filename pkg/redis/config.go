package redis

import (
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

// Config holds the configuration for the Redis client.
type Config struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	PoolTimeout  time.Duration
	Logger       zerolog.Logger
}

// DefaultConfig returns a Config with production-ready defaults.
// If the provided logger is disabled, a no-op logger is used instead.
func DefaultConfig(addr, password string, db int, logger zerolog.Logger) Config {
	if logger.GetLevel() == zerolog.Disabled {
		logger = zerolog.Nop()
	}

	return Config{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10 * runtime.NumCPU(),
		MinIdleConns: 5,
		MaxRetries:   5,
		PoolTimeout:  5 * time.Second,
		Logger:       logger,
	}
}
