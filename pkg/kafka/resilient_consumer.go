package kafka

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// ResilientConsumer wraps a Consumer with retry logic and DLQ support
type ResilientConsumer struct {
	consumer *Consumer
	dlq      *DeadLetterQueue
	logger   zerolog.Logger
	config   ConsumerConfig
}

// NewResilientConsumer creates a consumer with built-in retry and DLQ
func NewResilientConsumer(cfg ConsumerConfig, dlq *DeadLetterQueue) *ResilientConsumer {
	return &ResilientConsumer{
		consumer: NewConsumer(cfg),
		dlq:      dlq,
		logger:   cfg.Logger,
		config:   cfg,
	}
}

// NewResilientConsumerFromConsumer wraps an existing Consumer with retry and DLQ
func NewResilientConsumerFromConsumer(consumer *Consumer, dlq *DeadLetterQueue, logger zerolog.Logger) *ResilientConsumer {
	return &ResilientConsumer{
		consumer: consumer,
		dlq:      dlq,
		logger:   logger,
	}
}

// Start begins consuming with retry and DLQ handling
func (rc *ResilientConsumer) Start(ctx context.Context, handler MessageHandler) {
	wrappedHandler := rc.wrapHandlerWithRetry(handler)
	rc.consumer.Start(ctx, wrappedHandler)
}

// wrapHandlerWithRetry adds retry logic and DLQ to the handler
func (rc *ResilientConsumer) wrapHandlerWithRetry(handler MessageHandler) MessageHandler {
	return func(ctx context.Context, msg Message) error {
		err := handler(ctx, msg)
		if err == nil {
			return nil
		}

		// Check if we should retry
		if rc.dlq != nil && !rc.dlq.ShouldRetry(msg) {
			// Max retries reached, send to DLQ
			rc.logger.Warn().
				Str("event_type", msg.EventType).
				Str("saga_id", msg.SagaID).
				Int("retry_count", msg.RetryCount).
				Msg("max retries reached, sending to DLQ")

			if dlqErr := rc.dlq.SendToDLQ(ctx, msg, err.Error()); dlqErr != nil {
				rc.logger.Error().Err(dlqErr).Str("event_type", msg.EventType).
					Str("saga_id", msg.SagaID).Msg("failed to send message to DLQ")
			}
			return nil // Message handled (sent to DLQ)
		}

		// Retry with exponential backoff
		retryMsg := IncrementRetryCount(msg)
		backoff := calculateBackoff(retryMsg.RetryCount)

		rc.logger.Warn().
			Str("event_type", msg.EventType).
			Str("saga_id", msg.SagaID).
			Int("retry_count", retryMsg.RetryCount).
			Dur("backoff", backoff).
			Msg("retrying message processing")

		RecordRetry(msg.EventType, msg.EventType)

		time.Sleep(backoff)

		// Re-publish the message with incremented retry count for re-processing
		// In production, this could re-publish to the same topic or a retry topic
		return handler(ctx, retryMsg)
	}
}

// calculateBackoff computes exponential backoff with jitter
func calculateBackoff(retryCount int) time.Duration {
	base := time.Second
	maxBackoff := 30 * time.Second

	backoff := base
	for i := 0; i < retryCount; i++ {
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
			break
		}
	}

	// Add jitter (±20%)
	jitter := time.Duration(float64(backoff) * 0.2)
	return backoff + jitter
}

// Close gracefully closes the consumer and DLQ
func (rc *ResilientConsumer) Close() error {
	if err := rc.consumer.Close(); err != nil {
		return err
	}
	if rc.dlq != nil {
		return rc.dlq.Close()
	}
	return nil
}
