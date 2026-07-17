package kafka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(EventOrderCreated, "saga-123", "order-456", "order", map[string]string{"key": "value"})

	assert.NotEmpty(t, msg.EventID)
	assert.Equal(t, EventOrderCreated, msg.EventType)
	assert.Equal(t, "saga-123", msg.SagaID)
	assert.Equal(t, "order-456", msg.AggregateID)
	assert.Equal(t, "order", msg.AggregateType)
	assert.Equal(t, 0, msg.RetryCount)
	assert.WithinDuration(t, time.Now().UTC(), msg.Timestamp, time.Second)
}

func TestMessage_Serialize(t *testing.T) {
	msg := NewMessage(EventOrderCreated, "saga-123", "order-456", "order", map[string]string{"key": "value"})

	data, err := msg.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize and verify
	deserialized, err := DeserializeMessage(data)
	require.NoError(t, err)
	assert.Equal(t, msg.EventID, deserialized.EventID)
	assert.Equal(t, msg.EventType, deserialized.EventType)
	assert.Equal(t, msg.SagaID, deserialized.SagaID)
}

func TestMessage_WithCorrelationID(t *testing.T) {
	msg := NewMessage(EventOrderCreated, "saga-123", "order-456", "order", nil)
	msg = msg.WithCorrelationID("corr-789")

	assert.Equal(t, "corr-789", msg.CorrelationID)
}

func TestMessage_WithRetryCount(t *testing.T) {
	msg := NewMessage(EventOrderCreated, "saga-123", "order-456", "order", nil)
	msg = msg.WithRetryCount(3)

	assert.Equal(t, 3, msg.RetryCount)
}

func TestDeserializeMessage_Invalid(t *testing.T) {
	_, err := DeserializeMessage([]byte("invalid json"))
	assert.Error(t, err)
}

func TestIncrementRetryCount(t *testing.T) {
	msg := NewMessage(EventOrderCreated, "saga-123", "order-456", "order", nil)
	msg.RetryCount = 2

	incremented := IncrementRetryCount(msg)
	assert.Equal(t, 3, incremented.RetryCount)
}

func TestDeadLetterQueue_ShouldRetry(t *testing.T) {
	dlq := &DeadLetterQueue{config: DeadLetterConfig{MaxRetries: 3}}

	tests := []struct {
		name       string
		retryCount int
		expected   bool
	}{
		{"zero retries", 0, true},
		{"one retry", 1, true},
		{"two retries", 2, true},
		{"three retries (max)", 3, false},
		{"four retries", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{RetryCount: tt.retryCount}
			assert.Equal(t, tt.expected, dlq.ShouldRetry(msg))
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name       string
		retryCount int
		minBackoff time.Duration
		maxBackoff time.Duration
	}{
		{"first retry", 1, time.Second, 3*time.Second},
		{"second retry", 2, 2*time.Second, 6*time.Second},
		{"third retry", 3, 4*time.Second, 12*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateBackoff(tt.retryCount)
			assert.GreaterOrEqual(t, backoff, tt.minBackoff)
			assert.LessOrEqual(t, backoff, tt.maxBackoff)
		})
	}
}
