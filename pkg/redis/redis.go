package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Client provides a production-ready Redis client with connection pooling,
// timeouts, OpenTelemetry tracing, and structured logging.
type Client struct {
	client *redis.Client
	config Config
	logger zerolog.Logger
}

// NewClient creates a new Redis client with production-ready connection pooling,
// configurable timeouts, and retry policies.
// Parameters:
//   - cfg: Configuration for the Redis connection and client behavior.
//
// Returns:
//   - *Client: A fully initialized Redis client ready for use.
func NewClient(cfg Config) *Client {
	if cfg.Logger.GetLevel() == zerolog.Disabled {
		cfg.Logger = zerolog.Nop()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		PoolTimeout:  cfg.PoolTimeout,
	})

	return &Client{
		client: client,
		config: cfg,
		logger: cfg.Logger,
	}
}

// Close gracefully shuts down the Redis client and releases all connection pool resources.
// Parameters:
//   - ctx: Context for the close operation.
//
// Returns:
//   - error: Any error encountered during shutdown.
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping checks connectivity to the Redis server.
// Parameters:
//   - ctx: Context for the ping operation.
//
// Returns:
//   - error: Any error encountered during the ping.
func (c *Client) Ping(ctx context.Context) error {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.ping")
	defer span.End()

	err := c.client.Ping(ctx).Err()
	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Msg("redis: ping failed")
		return errors.Wrap(err, "redis: ping failed")
	}

	return nil
}

// Get retrieves the value for a given key.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to retrieve.
//
// Returns:
//   - string: The value stored at the key.
//   - error: Any error encountered during the operation.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.get")
	defer span.End()

	start := time.Now()
	val, err := c.client.Get(ctx, key).Result()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("key", key).Dur("latency", latency).Msg("redis: get failed")
		return "", errors.Wrap(err, "redis: get failed")
	}

	c.logger.Debug().Str("key", key).Dur("latency", latency).Msg("redis: get completed")
	return val, nil
}

// Set stores a key-value pair with an optional expiration duration.
// Pass 0 for expiration to store without a TTL.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to set.
//   - value: The value to store.
//   - expiration: TTL for the key. 0 means no expiration.
//
// Returns:
//   - error: Any error encountered during the operation.
func (c *Client) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.set")
	defer span.End()

	start := time.Now()
	err := c.client.Set(ctx, key, value, expiration).Err()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("key", key).Dur("latency", latency).Msg("redis: set failed")
		return errors.Wrap(err, "redis: set failed")
	}

	c.logger.Debug().Str("key", key).Dur("expiration", expiration).Dur("latency", latency).Msg("redis: set completed")
	return nil
}

// SetEX stores a key-value pair with a specific expiration duration.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to set.
//   - value: The value to store.
//   - expiration: TTL for the key.
//
// Returns:
//   - error: Any error encountered during the operation.
func (c *Client) SetEX(ctx context.Context, key, value string, expiration time.Duration) error {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.setex")
	defer span.End()

	start := time.Now()
	err := c.client.Set(ctx, key, value, expiration).Err()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("key", key).Dur("latency", latency).Msg("redis: setex failed")
		return errors.Wrap(err, "redis: setex failed")
	}

	c.logger.Debug().Str("key", key).Dur("expiration", expiration).Dur("latency", latency).Msg("redis: setex completed")
	return nil
}

// SetNX sets a key-value pair only if the key does not already exist.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to set.
//   - value: The value to store.
//   - expiration: TTL for the key. 0 means no expiration.
//
// Returns:
//   - bool: true if the key was set, false if it already existed.
//   - error: Any error encountered during the operation.
func (c *Client) SetNX(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.setnx")
	defer span.End()

	start := time.Now()
	ok, err := c.client.SetNX(ctx, key, value, expiration).Result()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("key", key).Dur("latency", latency).Msg("redis: setnx failed")
		return false, errors.Wrap(err, "redis: setnx failed")
	}

	c.logger.Debug().Str("key", key).Bool("set", ok).Dur("latency", latency).Msg("redis: setnx completed")
	return ok, nil
}

// Delete removes one or more keys.
// Parameters:
//   - ctx: Context for the operation.
//   - keys: One or more Redis keys to delete.
//
// Returns:
//   - int64: The number of keys that were removed.
//   - error: Any error encountered during the operation.
func (c *Client) Delete(ctx context.Context, keys ...string) (int64, error) {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.del")
	defer span.End()

	start := time.Now()
	val, err := c.client.Del(ctx, keys...).Result()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Interface("keys", keys).Dur("latency", latency).Msg("redis: del failed")
		return 0, errors.Wrap(err, "redis: del failed")
	}

	c.logger.Debug().Interface("keys", keys).Int64("removed", val).Dur("latency", latency).Msg("redis: del completed")
	return val, nil
}

// Exists checks whether one or more keys exist.
// Parameters:
//   - ctx: Context for the operation.
//   - keys: One or more Redis keys to check.
//
// Returns:
//   - int64: The number of keys that exist.
//   - error: Any error encountered during the operation.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	ctx, span := instrumentation.NewTraceSpan(ctx, "redis.exists")
	defer span.End()

	start := time.Now()
	val, err := c.client.Exists(ctx, keys...).Result()
	latency := time.Since(start)

	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Interface("keys", keys).Dur("latency", latency).Msg("redis: exists failed")
		return 0, errors.Wrap(err, "redis: exists failed")
	}

	c.logger.Debug().Interface("keys", keys).Int64("count", val).Dur("latency", latency).Msg("redis: exists completed")
	return val, nil
}

// SetJSON marshals a value to JSON and stores it in Redis.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to store the JSON value.
//   - value: The value to marshal and store.
//   - expiration: TTL for the key. 0 means no expiration.
//
// Returns:
//   - error: Any error encountered during marshaling or storage.
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "redis: json marshal failed")
	}

	return c.Set(ctx, key, string(data), expiration)
}

// GetJSON retrieves a value from Redis and unmarshals it from JSON.
// Parameters:
//   - ctx: Context for the operation.
//   - key: The Redis key to retrieve.
//   - dest: The destination to unmarshal the JSON value into.
//
// Returns:
//   - error: Any error encountered during retrieval or unmarshaling.
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return errors.Wrap(err, "redis: json unmarshal failed")
	}

	return nil
}
