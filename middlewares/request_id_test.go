package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dafailyasa/go-package/pkg/constant"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type RequestIDMiddlewareSuite struct {
	suite.Suite

	echo *echo.Echo
}

func TestRequestIDMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(RequestIDMiddlewareSuite))
}

func (s *RequestIDMiddlewareSuite) SetupTest() {
	s.echo = echo.New()
}

func (s *RequestIDMiddlewareSuite) TestRequestIDMiddleware_WithExistingRequestID() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(constant.XRequestID, "request-123")

	rec := httptest.NewRecorder()
	ctx := s.echo.NewContext(req, rec)

	called := false

	err := RequestIDMiddleware()(func(c echo.Context) error {
		called = true

		s.Equal("request-123", c.Request().Header.Get(constant.XRequestID))
		s.Equal("request-123", c.Response().Header().Get(constant.XRequestID))
		s.Equal("request-123", c.Get(constant.XRequestID))

		value := c.Request().Context().Value(constant.XRequestID)
		s.Equal("request-123", value)

		return c.NoContent(http.StatusOK)
	})(ctx)

	s.NoError(err)
	s.True(called)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *RequestIDMiddlewareSuite) TestRequestIDMiddleware_GenerateRequestID() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	ctx := s.echo.NewContext(req, rec)

	var generatedID string

	err := RequestIDMiddleware()(func(c echo.Context) error {
		generatedID = c.Request().Header.Get(constant.XRequestID)

		s.NotEmpty(generatedID)

		_, parseErr := uuid.Parse(generatedID)
		s.NoError(parseErr)

		s.Equal(generatedID, c.Response().Header().Get(constant.XRequestID))
		s.Equal(generatedID, c.Get(constant.XRequestID))
		s.Equal(generatedID, c.Request().Context().Value(constant.XRequestID))

		return c.NoContent(http.StatusOK)
	})(ctx)

	s.NoError(err)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *RequestIDMiddlewareSuite) TestRequestIDMiddleware_HandlerReturnsError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	ctx := s.echo.NewContext(req, rec)

	expectedErr := echo.NewHTTPError(http.StatusBadRequest, "bad request")

	err := RequestIDMiddleware()(func(c echo.Context) error {
		return expectedErr
	})(ctx)

	s.Error(err)
	s.Equal(expectedErr, err)

	// Middleware should still inject the request ID.
	requestID := ctx.Request().Header.Get(constant.XRequestID)

	s.NotEmpty(requestID)

	_, parseErr := uuid.Parse(requestID)
	s.NoError(parseErr)

	s.Equal(requestID, ctx.Response().Header().Get(constant.XRequestID))
}
