package redis

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		password string
		db       int
	}{
		{
			name:     "Positive - standard address",
			addr:     "localhost:6379",
			password: "",
			db:       0,
		},
		{
			name:     "Positive - empty address accepted",
			addr:     "",
			password: "",
			db:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig(tt.addr, tt.password, tt.db, zerolog.Nop())

			assert.Equal(t, tt.addr, cfg.Addr)
			assert.Equal(t, tt.password, cfg.Password)
			assert.Equal(t, tt.db, cfg.DB)
			assert.Equal(t, 5*time.Second, cfg.DialTimeout)
			assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
			assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
			assert.Equal(t, 10*runtime.NumCPU(), cfg.PoolSize)
			assert.Equal(t, 5, cfg.MinIdleConns)
			assert.Equal(t, 5, cfg.MaxRetries)
			assert.Equal(t, 5*time.Second, cfg.PoolTimeout)
		})
	}
}

func TestDefaultConfig_Timeouts(t *testing.T) {
	cfg := DefaultConfig("localhost:6379", "test", 0, zerolog.Nop())

	assert.Greater(t, cfg.DialTimeout, time.Duration(0), "dial timeout must be positive")
	assert.Greater(t, cfg.ReadTimeout, time.Duration(0), "read timeout must be positive")
	assert.Greater(t, cfg.WriteTimeout, time.Duration(0), "write timeout must be positive")
	assert.Greater(t, cfg.PoolTimeout, time.Duration(0), "pool timeout must be positive")
}

func TestDefaultConfig_PoolSettings(t *testing.T) {
	cfg := DefaultConfig("localhost:6379", "test", 0, zerolog.Nop())

	assert.Greater(t, cfg.PoolSize, 0, "pool size must be positive")
	assert.GreaterOrEqual(t, cfg.MinIdleConns, 0, "min idle conns must be non-negative")
	assert.Greater(t, cfg.MaxRetries, 0, "max retries must be positive")
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "Positive - default config",
			config: DefaultConfig("localhost:6379", "test", 0, zerolog.Nop()),
		},
		{
			name: "Positive - custom config",
			config: Config{
				Addr:         "redis-host:6380",
				Password:     "secret",
				DB:           3,
				DialTimeout:  10 * time.Second,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				PoolSize:     20,
				MinIdleConns: 10,
				MaxRetries:   3,
				PoolTimeout:  10 * time.Second,
				Logger:       zerolog.Nop(),
			},
		},
		{
			name: "Positive - disabled logger uses nop fallback",
			config: Config{
				Addr: "localhost:6379",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			require.NotNil(t, client)
			assert.NotNil(t, client.client)
		})
	}
}

func TestNewClient_NilLoggerFallback(t *testing.T) {
	cfg := Config{
		Addr: "localhost:6379",
	}

	client := NewClient(cfg)
	require.NotNil(t, client)
	assert.NotNil(t, client.logger)
}

func TestSetJSON_MarshalSuccess(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := testStruct{Name: "test", Value: 42}
	result, err := json.Marshal(data)
	require.NoError(t, err)
	assert.Contains(t, string(result), "test")
	assert.Contains(t, string(result), "42")
}

func TestSetJSON_MarshalFailure(t *testing.T) {
	client := NewClient(DefaultConfig("localhost:6379", "test", 0, zerolog.Nop()))
	require.NotNil(t, client)

	err := client.SetJSON(context.Background(), "key", make(chan int), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis: json marshal failed")
}

func TestGetJSON_UnmarshalFailure(t *testing.T) {
	client := NewClient(DefaultConfig("localhost:6379", "test", 0, zerolog.Nop()))
	require.NotNil(t, client)

	var dest int
	err := client.GetJSON(context.Background(), "key", dest)
	assert.Error(t, err)
}

func TestConfig_LoggerFallback(t *testing.T) {
	tests := []struct {
		name   string
		logger zerolog.Logger
	}{
		{
			name:   "Positive - nop logger",
			logger: zerolog.Nop(),
		},
		{
			name:   "Positive - disabled logger",
			logger: zerolog.Logger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig("localhost:6379", "test", 0, tt.logger)
			assert.NotNil(t, cfg.Logger)
		})
	}
}
