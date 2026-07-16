package utils

import (
	"context"

	"github.com/google/uuid"
)

const (
	XRequestID = "x-request-id"
)

func InjectRequestID(ctx context.Context, reqID, chainID, journeyID string) context.Context {
	if reqID == "" {
		reqID = uuid.New().String()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = injectToCtx(ctx, XRequestID, reqID)
	return ctx
}

func injectToCtx(ctx context.Context, key, value string) context.Context {
	if value != "" {
		return context.WithValue(ctx, key, value)
	}
	return ctx
}

func ExtractRequestID(ctx context.Context) string {
	if ctx == nil {
		return uuid.New().String()
	}
	if requestID, ok := ctx.Value(XRequestID).(string); ok {
		return requestID
	}
	return uuid.New().String()
}
