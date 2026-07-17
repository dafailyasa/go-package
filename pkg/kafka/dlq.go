package kafka

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// DeadLetterQueue handles messages that have failed processing after max retries
type DeadLetterQueue struct {
	producer *Producer
	config   DeadLetterConfig
	logger   zerolog.Logger
}

// NewDeadLetterQueue creates a new DLQ handler
func NewDeadLetterQueue(cfg DeadLetterConfig) *DeadLetterQueue {
	producerCfg := ProducerConfig{
		Brokers:      cfg.Brokers,
		Balancer:     "roundrobin",
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        false,
		Logger:       cfg.Logger,
	}

	return &DeadLetterQueue{
		producer: NewProducer(producerCfg),
		config:   cfg,
		logger:   cfg.Logger,
	}
}

// SendToDLQ publishes a failed message to the dead letter topic
func (dlq *DeadLetterQueue) SendToDLQ(ctx context.Context, originalMsg Message, reason string) error {
	dlqMsg := Message{
		EventID:        originalMsg.EventID,
		EventType:      originalMsg.EventType + ".dlq",
		SagaID:         originalMsg.SagaID,
		AggregateID:    originalMsg.AggregateID,
		AggregateType:  originalMsg.AggregateType,
		Payload:        originalMsg.Payload,
		Timestamp:      time.Now().UTC(),
		IdempotencyKey: originalMsg.IdempotencyKey,
		RetryCount:     originalMsg.RetryCount,
		CorrelationID:  originalMsg.CorrelationID,
	}

	dlq.logger.Warn().
		Str("topic", dlq.config.Topic).
		Str("event_type", originalMsg.EventType).
		Str("saga_id", originalMsg.SagaID).
		Str("event_id", originalMsg.EventID).
		Int("retry_count", originalMsg.RetryCount).
		Str("reason", reason).
		Msg("sending message to dead letter queue")

	return dlq.producer.Publish(ctx, dlq.config.Topic, dlqMsg)
}

// ShouldRetry determines if a message should be retried based on retry count
func (dlq *DeadLetterQueue) ShouldRetry(msg Message) bool {
	return msg.RetryCount < dlq.config.MaxRetries
}

// IncrementRetryCount creates a new message with incremented retry count
func IncrementRetryCount(msg Message) Message {
	msg.RetryCount++
	return msg
}

// Close closes the DLQ producer
func (dlq *DeadLetterQueue) Close() error {
	return dlq.producer.Close()
}
