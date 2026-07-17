package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type CacheWithRevalidationSuite struct {
	suite.Suite

	echo *echo.Echo
}

func TestCacheWithRevalidationSuite(t *testing.T) {
	suite.Run(t, new(CacheWithRevalidationSuite))
}

func (s *CacheWithRevalidationSuite) SetupTest() {
	s.echo = echo.New()
}

func (s *CacheWithRevalidationSuite) TestCacheWithRevalidation_Success() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	ctx := s.echo.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	err := CacheWithRevalidation(handler)(ctx)

	s.NoError(err)
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("success", rec.Body.String())

	headers := rec.Header()

	s.Equal(
		"no-cache, max-age=120, must-revalidate",
		headers.Get("Cache-Control"),
	)

	s.NotEmpty(headers.Get("Expires"))
	s.NotEmpty(headers.Get("Last-Modified"))

	_, err = time.Parse(http.TimeFormat, headers.Get("Expires"))
	s.NoError(err)

	_, err = time.Parse(http.TimeFormat, headers.Get("Last-Modified"))
	s.NoError(err)
}

func (s *CacheWithRevalidationSuite) TestCacheWithRevalidation_HandlerError() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	ctx := s.echo.NewContext(req, rec)

	expectedErr := echo.NewHTTPError(http.StatusBadRequest, "bad request")

	handler := func(c echo.Context) error {
		return expectedErr
	}

	err := CacheWithRevalidation(handler)(ctx)

	s.Error(err)
	s.Equal(expectedErr, err)

	headers := rec.Header()

	s.Equal(
		"no-cache, max-age=120, must-revalidate",
		headers.Get("Cache-Control"),
	)

	s.NotEmpty(headers.Get("Expires"))
	s.NotEmpty(headers.Get("Last-Modified"))
}

func (s *CacheWithRevalidationSuite) TestCacheWithRevalidation_HeadersExist() {
	req := httptest.NewRequest(http.MethodGet, "/cache", nil)
	rec := httptest.NewRecorder()

	ctx := s.echo.NewContext(req, rec)

	err := CacheWithRevalidation(func(c echo.Context) error {
		return nil
	})(ctx)

	s.NoError(err)

	headers := rec.Header()

	s.Contains(headers, "Cache-Control")
	s.Contains(headers, "Expires")
	s.Contains(headers, "Last-Modified")
}
