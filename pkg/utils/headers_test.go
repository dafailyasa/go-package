package utils

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInjectRequestID(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		requestID  string
		expectedID string
		expectUUID bool
	}{
		{
			name:       "Positive - Existing Request ID",
			ctx:        context.Background(),
			requestID:  "req-123",
			expectedID: "req-123",
		},
		{
			name:       "Positive - Nil Context",
			ctx:        nil,
			requestID:  "req-456",
			expectedID: "req-456",
		},
		{
			name:       "Negative - Empty Request ID Generates UUID",
			ctx:        context.Background(),
			requestID:  "",
			expectUUID: true,
		},
		{
			name:       "Negative - Nil Context And Empty Request ID",
			ctx:        nil,
			requestID:  "",
			expectUUID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := InjectRequestID(tt.ctx, tt.requestID, "", "")

			id := ctx.Value(XRequestID).(string)

			if tt.expectUUID {
				_, err := uuid.Parse(id)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestExtractRequestID(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		expectedID string
		expectUUID bool
	}{
		{
			name:       "Positive - Request ID Exists",
			ctx:        context.WithValue(context.Background(), XRequestID, "request-123"),
			expectedID: "request-123",
		},
		{
			name:       "Negative - Nil Context",
			ctx:        nil,
			expectUUID: true,
		},
		{
			name:       "Negative - Missing Request ID",
			ctx:        context.Background(),
			expectUUID: true,
		},
		{
			name:       "Negative - Invalid Type",
			ctx:        context.WithValue(context.Background(), XRequestID, 123),
			expectUUID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := ExtractRequestID(tt.ctx)

			if tt.expectUUID {
				_, err := uuid.Parse(id)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestInjectToCtx(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected interface{}
	}{
		{
			name:     "Positive - Value Exists",
			value:    "abc123",
			expected: "abc123",
		},
		{
			name:     "Negative - Empty Value",
			value:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := injectToCtx(context.Background(), XRequestID, tt.value)

			assert.Equal(t, tt.expected, ctx.Value(XRequestID))
		})
	}
}
