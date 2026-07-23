package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// EventType constants for the e-commerce saga
const (
	// Order events
	EventOrderCreated   = "order.created"
	EventOrderConfirmed = "order.confirmed"
	EventOrderFailed    = "order.failed"
	EventOrderCancelled = "order.cancelled"

	// Inventory events
	EventStockReserved          = "inventory.stock_reserved"
	EventStockReservationFailed = "inventory.stock_reservation_failed"
	EventStockReleased          = "inventory.stock_released"

	// Payment events
	EventPaymentProcessed = "payment.processed"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"

	// Shipping events
	EventShipmentCreated   = "shipping.created"
	EventShipmentFailed    = "shipping.failed"
	EventShipmentDelivered = "shipping.delivered"

	// Notification events
	EventNotificationSent   = "notification.sent"
	EventNotificationFailed = "notification.failed"
)

// Topic constants
const (
	TopicOrder        = "order.events"
	TopicInventory    = "inventory.events"
	TopicPayment      = "payment.events"
	TopicShipping     = "shipping.events"
	TopicNotification = "notification.events"
	TopicSaga         = "saga.events"
	TopicDeadLetter   = "saga.dead-letter"
)

// Message represents the envelope for all Kafka events
type Message struct {
	EventID        string      `json:"event_id"`
	EventType      string      `json:"event_type"`
	SagaID         string      `json:"saga_id"`
	AggregateID    string      `json:"aggregate_id"`
	AggregateType  string      `json:"aggregate_type"`
	Payload        interface{} `json:"payload"`
	Timestamp      time.Time   `json:"timestamp"`
	IdempotencyKey string      `json:"idempotency_key"`
	RetryCount     int         `json:"retry_count"`
	CorrelationID  string      `json:"correlation_id,omitempty"`
}

// NewMessage creates a new message with auto-generated IDs
func NewMessage(eventType, sagaID, aggregateID, aggregateType string, payload interface{}) Message {
	return Message{
		EventID:        uuid.New().String(),
		EventType:      eventType,
		SagaID:         sagaID,
		AggregateID:    aggregateID,
		AggregateType:  aggregateType,
		Payload:        payload,
		Timestamp:      time.Now().UTC(),
		IdempotencyKey: uuid.New().String(),
		RetryCount:     0,
	}
}

// WithCorrelationID adds a correlation ID to the message
func (m Message) WithCorrelationID(correlationID string) Message {
	m.CorrelationID = correlationID
	return m
}

// WithRetryCount sets the retry count on the message
func (m Message) WithRetryCount(count int) Message {
	m.RetryCount = count
	return m
}

// Serialize converts the message to JSON bytes
func (m Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage converts JSON bytes back to a Message
func DeserializeMessage(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}

// MessageHandler is a function that processes a received message
type MessageHandler func(ctx context.Context, msg Message) error

// ProducerConfig holds configuration for the Kafka producer
type ProducerConfig struct {
	Brokers      []string
	Balancer     string // "roundrobin", "leastbytes", "hash"
	BatchTimeout time.Duration
	BatchSize    int
	Async        bool
	Logger       zerolog.Logger
}

// ConsumerConfig holds configuration for the Kafka consumer
type ConsumerConfig struct {
	Brokers     []string
	GroupID     string
	Topics      []string
	MinBytes    int
	MaxBytes    int
	MaxWait     time.Duration
	StartOffset int64 // -1 for earliest, -2 for latest
	Logger      zerolog.Logger
}

// DeadLetterConfig holds configuration for the DLQ
type DeadLetterConfig struct {
	Brokers    []string
	Topic      string
	MaxRetries int
	Logger     zerolog.Logger
}

// DefaultProducerConfig returns a producer config with sensible defaults
func DefaultProducerConfig(brokers []string, logger zerolog.Logger) ProducerConfig {
	return ProducerConfig{
		Brokers:      brokers,
		Balancer:     "roundrobin",
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        false,
		Logger:       logger,
	}
}

// DefaultConsumerConfig returns a consumer config with sensible defaults
func DefaultConsumerConfig(brokers []string, groupID string, topics []string, logger zerolog.Logger) ConsumerConfig {
	return ConsumerConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topics:      topics,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
		MaxWait:     5 * time.Second,
		StartOffset: -1, // earliest
		Logger:      logger,
	}
}

// DefaultDeadLetterConfig returns a DLQ config with sensible defaults
func DefaultDeadLetterConfig(brokers []string, logger zerolog.Logger) DeadLetterConfig {
	return DeadLetterConfig{
		Brokers:    brokers,
		Topic:      TopicDeadLetter,
		MaxRetries: 3,
		Logger:     logger,
	}
}
