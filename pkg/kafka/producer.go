package kafka

import (
	"context"
	"time"

	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"go.opentelemetry.io/otel/propagation"
)

// Producer handles publishing messages to Kafka topics
type Producer struct {
	writer      *kafka.Writer
	logger      zerolog.Logger
	propagator  propagation.TextMapPropagator
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig) *Producer {
	var balancer kafka.Balancer
	switch cfg.Balancer {
	case "leastbytes":
		balancer = &kafka.LeastBytes{}
	case "hash":
		balancer = &kafka.Hash{}
	default:
		balancer = &kafka.RoundRobin{}
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     balancer,
		BatchTimeout: cfg.BatchTimeout,
		BatchSize:    cfg.BatchSize,
		Async:        cfg.Async,
		RequiredAcks: kafka.RequireAll,
		Compression:  compress.Snappy,
		Logger:       kafka.LoggerFunc(cfg.Logger.Printf),
		ErrorLogger:  kafka.LoggerFunc(cfg.Logger.Printf),
	}

	return &Producer{
		writer:     writer,
		logger:     cfg.Logger,
		propagator: propagation.TraceContext{},
	}
}

// buildHeaders creates the base headers for a message
func (p *Producer) buildHeaders(msg Message) []kafka.Header {
	return []kafka.Header{
		{Key: "event_type", Value: []byte(msg.EventType)},
		{Key: "saga_id", Value: []byte(msg.SagaID)},
		{Key: "event_id", Value: []byte(msg.EventID)},
		{Key: "idempotency_key", Value: []byte(msg.IdempotencyKey)},
		{Key: "timestamp", Value: []byte(msg.Timestamp.Format(time.RFC3339))},
	}
}

// Publish sends a message to the specified topic with trace context propagation
func (p *Producer) Publish(ctx context.Context, topic string, msg Message) error {
	ctx, span := instrumentation.NewTraceSpan(ctx, "kafka.produce."+topic)
	defer span.End()

	data, err := msg.Serialize()
	if err != nil {
		instrumentation.RecordSpanError(span, err)
		return err
	}

	// Build headers and inject trace context
	headers := p.buildHeaders(msg)
	carrier := NewKafkaHeaderCarrier(headers)
	p.propagator.Inject(ctx, carrier)

	kafkaMsg := kafka.Message{
		Key:     []byte(msg.AggregateID),
		Value:   data,
		Headers: carrier.Headers(),
	}

	err = p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		instrumentation.RecordSpanError(span, err)
		p.logger.Error().Err(err).Str("topic", topic).Str("event_type", msg.EventType).
			Str("saga_id", msg.SagaID).Msg("failed to publish message")
		return err
	}

	p.logger.Info().Str("topic", topic).Str("event_type", msg.EventType).
		Str("saga_id", msg.SagaID).Str("event_id", msg.EventID).
		Msg("message published successfully")
	return nil
}

// PublishAsync sends a message asynchronously (fire and forget) with trace context
func (p *Producer) PublishAsync(ctx context.Context, topic string, msg Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}

	// Build headers and inject trace context
	headers := p.buildHeaders(msg)
	carrier := NewKafkaHeaderCarrier(headers)
	p.propagator.Inject(ctx, carrier)

	kafkaMsg := kafka.Message{
		Key:     []byte(msg.AggregateID),
		Value:   data,
		Headers: carrier.Headers(),
	}

	return p.writer.WriteMessages(ctx, kafkaMsg)
}

// Close gracefully closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
