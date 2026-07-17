package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LoggingMiddlewareSuite struct {
	suite.Suite

	echo      *echo.Echo
	logger    *zerolog.Logger
	logBuffer *bytes.Buffer
}

func TestLoggingMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(LoggingMiddlewareSuite))
}

func (s *LoggingMiddlewareSuite) SetupTest() {
	s.echo = echo.New()

	s.logBuffer = new(bytes.Buffer)
	logger := zerolog.New(s.logBuffer)
	s.logger = &logger
}

func (s *LoggingMiddlewareSuite) TestLoggingMiddleware_Success() {
	req := httptest.NewRequest(
		http.MethodPost,
		"/users?id=1",
		bytes.NewBufferString(`{"name":"john"}`),
	)

	req = req.WithContext(context.WithValue(
		req.Context(),
		"x-request-id",
		"req-123",
	))

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	handler := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)

		require.NoError(s.T(), err)
		assert.JSONEq(s.T(), `{"name":"john"}`, string(body))

		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}

	err := LoggingMiddleware(s.logger)(handler)(c)

	require.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusOK, rec.Code)
	assert.Contains(s.T(), s.logBuffer.String(), "[request]")
	assert.Contains(s.T(), s.logBuffer.String(), "[response]")
	assert.Contains(s.T(), s.logBuffer.String(), "req-123")
}

func (s *LoggingMiddlewareSuite) TestLoggingMiddleware_EmptyBody() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	c := s.echo.NewContext(req, rec)

	err := LoggingMiddleware(s.logger)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, rec.Code)
}

func (s *LoggingMiddlewareSuite) TestLoggingMiddleware_IgnoredPath() {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	c := s.echo.NewContext(req, rec)

	err := LoggingMiddleware(s.logger)(func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})(c)

	require.NoError(s.T(), err)

	assert.Empty(s.T(), s.logBuffer.String())
}

func (s *LoggingMiddlewareSuite) TestLoggingMiddleware_HandlerError() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	c := s.echo.NewContext(req, rec)

	expected := echo.NewHTTPError(http.StatusBadRequest, "bad request")

	err := LoggingMiddleware(s.logger)(func(c echo.Context) error {
		return expected
	})(c)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), expected, err)
}

func (s *LoggingMiddlewareSuite) TestCustomResponseWriter_Write() {
	rec := httptest.NewRecorder()

	buf := new(bytes.Buffer)

	writer := &CustomResponseWriter{
		ResponseWriter: rec,
		buf:            buf,
	}

	n, err := writer.Write([]byte("hello"))

	require.NoError(s.T(), err)
	assert.Equal(s.T(), 5, n)
	assert.Equal(s.T(), "hello", buf.String())
	assert.Equal(s.T(), "hello", rec.Body.String())
}

func (s *LoggingMiddlewareSuite) TestIsPathIgnored() {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Swagger", "/swagger/index.html", true},
		{"Swagger Prefix", "/api/swagger", true},
		{"Ping", "/ping", true},
		{"Metrics", "/metrics", true},
		{"User", "/users", false},
		{"Order", "/orders", false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			assert.Equal(s.T(), tt.want, isPathIgnored(tt.path))
		})
	}
}
