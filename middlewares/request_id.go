package middleware

import (
	"context"

	"github.com/dafailyasa/go-package/pkg/constant"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			requestID := req.Header.Get(constant.XRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			req.Header.Set(constant.XRequestID, requestID)

			res.Header().Set(constant.XRequestID, requestID)

			c.Set(constant.XRequestID, requestID)

			ctx := context.WithValue(req.Context(), constant.XRequestID, requestID)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}
