package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	XRequestID = "X-Request-ID"
)

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			requestID := req.Header.Get(XRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			req.Header.Set(XRequestID, requestID)

			res.Header().Set(XRequestID, requestID)

			c.Set(XRequestID, requestID)

			ctx := context.WithValue(req.Context(), XRequestID, requestID)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}
