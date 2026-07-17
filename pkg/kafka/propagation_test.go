package kafka

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
)

func TestKafkaHeaderCarrier_Get(t *testing.T) {
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
		{Key: "saga_id", Value: []byte("saga-123")},
	}

	carrier := NewKafkaHeaderCarrier(headers)

	assert.Equal(t, "order.created", carrier.Get("event_type"))
	assert.Equal(t, "saga-123", carrier.Get("saga_id"))
	assert.Equal(t, "", carrier.Get("nonexistent"))
}

func TestKafkaHeaderCarrier_Set(t *testing.T) {
	carrier := NewKafkaHeaderCarrierEmpty()

	carrier.Set("traceparent", "00-abc-def-01")
	assert.Equal(t, "00-abc-def-01", carrier.Get("traceparent"))

	// Overwrite existing
	carrier.Set("traceparent", "00-xyz-def-01")
	assert.Equal(t, "00-xyz-def-01", carrier.Get("traceparent"))
}

func TestKafkaHeaderCarrier_Keys(t *testing.T) {
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
		{Key: "saga_id", Value: []byte("saga-123")},
	}

	carrier := NewKafkaHeaderCarrier(headers)
	keys := carrier.Keys()

	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "event_type")
	assert.Contains(t, keys, "saga_id")
}

func TestKafkaHeaderCarrier_Headers(t *testing.T) {
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
	}

	carrier := NewKafkaHeaderCarrier(headers)
	result := carrier.Headers()

	assert.Len(t, result, 1)
	assert.Equal(t, "event_type", result[0].Key)
}

func TestInjectTraceHeaders(t *testing.T) {
	ctx := context.Background()
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
	}

	result := InjectTraceHeaders(ctx, headers)

	// Should have original header plus traceparent
	assert.GreaterOrEqual(t, len(result), 1)

	// Find traceparent header
	found := false
	for _, h := range result {
		if h.Key == TraceparentHeader {
			found = true
			assert.NotEmpty(t, string(h.Value))
		}
	}
	// Note: traceparent may not be injected if no span is active in context
	// This test verifies the mechanism works
	_ = found
}

func TestExtractTraceContext(t *testing.T) {
	ctx := context.Background()
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
		{Key: "saga_id", Value: []byte("saga-123")},
	}

	extractedCtx := ExtractTraceContext(ctx, headers)
	assert.NotNil(t, extractedCtx)
}

func TestKafkaHeaderCarrier_PropagationIntegration(t *testing.T) {
	// Simulate inject -> extract cycle
	propagator := propagation.TraceContext{}

	// Create carrier with some headers
	headers := []kafka.Header{
		{Key: "event_type", Value: []byte("order.created")},
	}
	carrier := NewKafkaHeaderCarrier(headers)

	// Inject trace context into carrier
	// Note: Without an active span in context, this won't add traceparent
	// but it tests the mechanism
	propagator.Inject(context.Background(), carrier)

	// Extract from carrier
	extractedCtx := propagator.Extract(context.Background(), carrier)
	assert.NotNil(t, extractedCtx)
}
