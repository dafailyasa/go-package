package middleware

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/dafailyasa/go-package/pkg/constant"
	"github.com/labstack/echo/v4"
)

type UserInfo struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
}

const UserKeyHeader = "user"

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			userInfoHeader := req.Header.Get("X-Userinfo")
			if userInfoHeader == "" {
				return apperror.Unauthorized(
					errors.New("You are not authorized to do this request missing X-Userinfo"),
				)
			}

			var userInfo UserInfo
			if err := json.Unmarshal([]byte(userInfoHeader), &userInfo); err != nil {
				return apperror.Unauthorized(
					errors.New("You are not authorized to do this request invalid X-Userinfo"),
				)
			}

			c.Set(UserKeyHeader, userInfo)
			ctx := context.WithValue(req.Context(), constant.XRequestID, userInfo)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

func GetUserInfoHeader(ctx echo.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Get(UserKeyHeader).(UserInfo)
	if !ok {
		return nil, false
	}

	return &userInfo, true
}
