package kafka

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "github.com/dafailyasa/go-package/kafka"

var (
	messagesConsumed metric.Int64Counter
	messagesProduced metric.Int64Counter
	dlqMessages      metric.Int64Counter
	retryAttempts    metric.Int64Counter
)

func init() {
	meter := otel.Meter(meterName)

	messagesConsumed, _ = meter.Int64Counter(
		"kafka.messages.consumed",
		metric.WithDescription("Total number of Kafka messages consumed"),
	)

	messagesProduced, _ = meter.Int64Counter(
		"kafka.messages.produced",
		metric.WithDescription("Total number of Kafka messages produced"),
	)

	dlqMessages, _ = meter.Int64Counter(
		"kafka.dlq.messages",
		metric.WithDescription("Total number of messages sent to dead letter queue"),
	)

	retryAttempts, _ = meter.Int64Counter(
		"kafka.retries",
		metric.WithDescription("Total number of message retry attempts"),
	)
}

func RecordMessageConsumed(topic, eventType string) {
	messagesConsumed.Add(nil, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("event_type", eventType),
		),
	)
}

func RecordMessageProduced(topic, eventType string) {
	messagesProduced.Add(nil, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("event_type", eventType),
		),
	)
}

func RecordDLQMessage(topic, eventType, reason string) {
	dlqMessages.Add(nil, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("event_type", eventType),
			attribute.String("reason", reason),
		),
	)
}

func RecordRetry(topic, eventType string) {
	retryAttempts.Add(nil, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
			attribute.String("event_type", eventType),
		),
	)
}
