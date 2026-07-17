package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/dafailyasa/go-package/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type CustomResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (crw *CustomResponseWriter) Write(b []byte) (int, error) {
	// Write to both the buffer and the original writer
	n, err := crw.buf.Write(b)
	if err != nil {
		return n, err
	}
	return crw.ResponseWriter.Write(b)
}

var ignoredPatterns = []*regexp.Regexp{}

func init() {
	// Define the dynamic parts of the paths
	dynamicPaths := []string{"swagger", "ping", "metrics"}

	for _, path := range dynamicPaths {
		// Create a pattern that allows for any prefix before the dynamic path
		pattern := fmt.Sprintf(`.*/%s`, path)
		ignoredPatterns = append(ignoredPatterns, regexp.MustCompile(pattern))
	}
	// Additional static pattern for /swagger/*
	ignoredPatterns = append(ignoredPatterns, regexp.MustCompile(`^/swagger/`))
}

func isPathIgnored(path string) bool {
	for _, pattern := range ignoredPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

func LoggingMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Track the start time of the request
			startTime := time.Now()
			ignoreLogging := isPathIgnored(req.URL.Path)
			var requestBodyJSON interface{}
			var sanitized any
			if req.Body != nil {
				// Read and restore the request body
				reqBody, err := io.ReadAll(req.Body)
				if err == nil {
					req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

					// Parse the request body as JSON
					if err := json.Unmarshal(reqBody, &requestBodyJSON); err != nil {
						requestBodyJSON = nil
					}

					sanitized = utils.SanitizeBodyParsed(reqBody)
				}
			}

			// Capture query parameters
			queryParams := req.URL.Query()

			// Create a custom response writer to capture the response body
			buf := new(bytes.Buffer)
			crw := &CustomResponseWriter{ResponseWriter: c.Response().Writer, buf: buf}
			c.Response().Writer = crw

			// Continue to the next middleware/handler
			err := next(c)

			if !ignoreLogging {
				xRequestID := utils.ExtractRequestID(c.Request().Context())
				logger.Info().
					Ctx(c.Request().Context()).
					Str("x-request-id", xRequestID).
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Str("ip", req.RemoteAddr).
					Interface("body", sanitized).
					Interface("query", queryParams).
					Msg("[request]")

				// Log response details including body using zerolog
				logger.Info().
					Ctx(c.Request().Context()).
					Str("x-request-id", xRequestID).
					Int("status", res.Status).
					Int64("size", res.Size).
					Dur("duration", time.Since(startTime)).
					Interface("body", utils.SanitizeBodyParsed(buf.Bytes())).
					Msg("[response]")
			}

			return err
		}
	}
}
