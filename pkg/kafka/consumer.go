package kafka

import (
	"context"
	"time"

	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// Consumer handles consuming messages from Kafka topics
type Consumer struct {
	reader     *kafka.Reader
	logger     zerolog.Logger
	config     ConsumerConfig
	propagator propagation.TextMapPropagator
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topics[0],
		MinBytes:    cfg.MinBytes,
		MaxBytes:    cfg.MaxBytes,
		MaxWait:     cfg.MaxWait,
		StartOffset: cfg.StartOffset,
		Logger:      kafka.LoggerFunc(cfg.Logger.Printf),
		ErrorLogger: kafka.LoggerFunc(cfg.Logger.Printf),
	})

	return &Consumer{
		reader:     reader,
		logger:     cfg.Logger,
		config:     cfg,
		propagator: propagation.TraceContext{},
	}
}

// NewMultiTopicConsumer creates multiple consumers, one per topic
func NewMultiTopicConsumer(cfg ConsumerConfig) []*Consumer {
	consumers := make([]*Consumer, 0, len(cfg.Topics))
	for _, topic := range cfg.Topics {
		topicCfg := cfg
		topicCfg.Topics = []string{topic}
		consumers = append(consumers, NewConsumer(topicCfg))
	}
	return consumers
}

// Config returns the consumer configuration
func (c *Consumer) Config() ConsumerConfig {
	return c.config
}

// Start begins consuming messages and processing them with the handler
func (c *Consumer) Start(ctx context.Context, handler MessageHandler) {
	c.logger.Info().Str("topic", c.config.Topics[0]).Str("group_id", c.config.GroupID).
		Msg("starting kafka consumer")

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info().Str("topic", c.config.Topics[0]).Msg("consumer context cancelled, stopping")
				return
			default:
				if err := c.consumeMessage(ctx, handler); err != nil {
					c.logger.Error().Err(err).Str("topic", c.config.Topics[0]).
						Msg("error consuming message, retrying...")
					time.Sleep(time.Second)
				}
			}
		}
	}()
}

// consumeMessage reads and processes a single message with trace context extraction
func (c *Consumer) consumeMessage(ctx context.Context, handler MessageHandler) error {
	kafkaMsg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return err
	}

	// Extract trace context from Kafka headers
	carrier := NewKafkaHeaderCarrier(kafkaMsg.Headers)
	extractedCtx := c.propagator.Extract(ctx, carrier)

	// Create a new span from the extracted context
	ctx, span := instrumentation.NewTraceSpan(extractedCtx, "kafka.consume."+c.config.Topics[0])
	defer span.End()

	// Add business attributes to span
	for _, header := range kafkaMsg.Headers {
		switch header.Key {
		case "saga_id":
			span.SetAttributes(attribute.String("saga_id", string(header.Value)))
		case "event_type":
			span.SetAttributes(attribute.String("event_type", string(header.Value)))
		case "event_id":
			span.SetAttributes(attribute.String("event_id", string(header.Value)))
		}
	}

	// Deserialize the message
	msg, err := DeserializeMessage(kafkaMsg.Value)
	if err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("topic", c.config.Topics[0]).
			Msg("failed to deserialize message")
		return err
	}

	c.logger.Info().Str("topic", c.config.Topics[0]).Str("event_type", msg.EventType).
		Str("saga_id", msg.SagaID).Str("event_id", msg.EventID).
		Int("retry_count", msg.RetryCount).
		Msg("received message")

	// Process the message
	if err := handler(ctx, msg); err != nil {
		instrumentation.RecordSpanError(span, err)
		c.logger.Error().Err(err).Str("topic", c.config.Topics[0]).
			Str("event_type", msg.EventType).Str("saga_id", msg.SagaID).
			Msg("failed to process message")
		return err
	}

	c.logger.Info().Str("topic", c.config.Topics[0]).Str("event_type", msg.EventType).
		Str("saga_id", msg.SagaID).Str("event_id", msg.EventID).
		Msg("message processed successfully")
	return nil
}

// Close gracefully closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
