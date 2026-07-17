package utils

import (
	"context"
	"testing"

	"github.com/dafailyasa/go-package/pkg/constant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RequestIDSuite struct {
	suite.Suite
}

func TestRequestIDSuite(t *testing.T) {
	suite.Run(t, new(RequestIDSuite))
}

func (s *RequestIDSuite) TestInjectRequestID() {
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
		s.Run(tt.name, func() {
			ctx := InjectRequestID(tt.ctx, tt.requestID, "", "")

			id, ok := ctx.Value(constant.XRequestID).(string)
			s.True(ok)

			if tt.expectUUID {
				_, err := uuid.Parse(id)
				s.NoError(err)
			} else {
				s.Equal(tt.expectedID, id)
			}
		})
	}
}

func (s *RequestIDSuite) TestExtractRequestID() {
	tests := []struct {
		name       string
		ctx        context.Context
		expectedID string
		expectUUID bool
	}{
		{
			name:       "Positive - Request ID Exists",
			ctx:        context.WithValue(context.Background(), constant.XRequestID, "request-123"),
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
			ctx:        context.WithValue(context.Background(), constant.XRequestID, 123),
			expectUUID: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			id := ExtractRequestID(tt.ctx)

			if tt.expectUUID {
				_, err := uuid.Parse(id)
				s.NoError(err)
			} else {
				s.Equal(tt.expectedID, id)
			}
		})
	}
}

func (s *RequestIDSuite) TestInjectToCtx() {
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
		s.Run(tt.name, func() {
			ctx := injectToCtx(context.Background(), constant.XRequestID, tt.value)

			s.Equal(tt.expected, ctx.Value(constant.XRequestID))
		})
	}
}
