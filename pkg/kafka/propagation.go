package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

const (
	// TraceparentHeader is the W3C Trace Context traceparent header key
	TraceparentHeader = "traceparent"
	// TracestateHeader is the W3C Trace Context tracestate header key
	TracestateHeader = "tracestate"
)

// KafkaHeaderCarrier implements propagation.TextMapCarrier for kafka headers
type KafkaHeaderCarrier struct {
	headers []kafka.Header
}

// NewKafkaHeaderCarrier creates a carrier from existing headers
func NewKafkaHeaderCarrier(headers []kafka.Header) *KafkaHeaderCarrier {
	return &KafkaHeaderCarrier{headers: headers}
}

// NewKafkaHeaderCarrierEmpty creates an empty carrier
func NewKafkaHeaderCarrierEmpty() *KafkaHeaderCarrier {
	return &KafkaHeaderCarrier{headers: make([]kafka.Header, 0)}
}

// Get returns the value associated with the passed key
func (c *KafkaHeaderCarrier) Get(key string) string {
	for _, h := range c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set stores the key-value pair
func (c *KafkaHeaderCarrier) Set(key, value string) {
	for i, h := range c.headers {
		if h.Key == key {
			c.headers[i].Value = []byte(value)
			return
		}
	}
	c.headers = append(c.headers, kafka.Header{Key: key, Value: []byte(value)})
}

// Keys lists the keys stored in this carrier
func (c *KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c.headers))
	for i, h := range c.headers {
		keys[i] = h.Key
	}
	return keys
}

// Headers returns the underlying kafka headers
func (c *KafkaHeaderCarrier) Headers() []kafka.Header {
	return c.headers
}

// InjectTraceHeaders injects the current trace context into Kafka message headers
func InjectTraceHeaders(ctx context.Context, headers []kafka.Header) []kafka.Header {
	carrier := NewKafkaHeaderCarrier(headers)
	propagation.TraceContext{}.Inject(ctx, carrier)
	return carrier.Headers()
}

// ExtractTraceContext extracts trace context from Kafka message headers and returns an enriched context
func ExtractTraceContext(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := NewKafkaHeaderCarrier(headers)
	return propagation.TraceContext{}.Extract(ctx, carrier)
}
