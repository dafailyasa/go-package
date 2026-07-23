package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareSuite struct {
	suite.Suite

	echo *echo.Echo
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareSuite))
}

func (s *AuthMiddlewareSuite) SetupTest() {
	s.echo = echo.New()
}

func (s *AuthMiddlewareSuite) TestAuthMiddleware() {
	tests := []struct {
		name        string
		header      string
		expectError bool
		expectUser  *UserInfo
	}{
		{
			name:   "success",
			header: `{"sub":"user-123","role":"admin"}`,
			expectUser: &UserInfo{
				Sub:  "user-123",
				Role: "admin",
			},
		},
		{
			name:        "missing header",
			expectError: true,
		},
		{
			name:        "invalid json",
			header:      "{invalid-json}",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.header != "" {
				req.Header.Set("X-Userinfo", tt.header)
			}

			rec := httptest.NewRecorder()
			ctx := s.echo.NewContext(req, rec)

			nextCalled := false

			handler := AuthMiddleware()(func(c echo.Context) error {
				nextCalled = true

				user, ok := GetUserInfoHeader(c)
				assert.True(t, ok)
				assert.Equal(t, tt.expectUser, user)

				return c.NoContent(http.StatusOK)
			})

			err := handler(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, nextCalled)

				user, ok := GetUserInfoHeader(ctx)
				assert.False(t, ok)
				assert.Nil(t, user)

				return
			}

			assert.NoError(t, err)
			assert.True(t, nextCalled)

			user, ok := GetUserInfoHeader(ctx)
			assert.True(t, ok)
			assert.Equal(t, tt.expectUser, user)
		})
	}
}

func (s *AuthMiddlewareSuite) TestGetUserInfoHeader() {
	s.T().Run("found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		ctx := s.echo.NewContext(req, rec)

		expected := UserInfo{
			Sub:  "user-123",
			Role: "admin",
		}

		ctx.Set(UserKeyHeader, expected)

		actual, ok := GetUserInfoHeader(ctx)

		assert.True(t, ok)
		assert.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	s.T().Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		ctx := s.echo.NewContext(req, rec)

		actual, ok := GetUserInfoHeader(ctx)

		assert.False(t, ok)
		assert.Nil(t, actual)
	})
}
