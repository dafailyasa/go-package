# go-package

[![Go Reference](https://pkg.go.dev/badge/github.com/dafailyasa/go-package.svg)](https://pkg.go.dev/github.com/dafailyasa/go-package)
[![Go Version](https://img.shields.io/badge/Go-1.26.1+-00ADD8.svg?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`go-package` is a comprehensive, production-ready utility toolkit designed to streamline development in Go. It provides unified, high-performance implementations for common backend requirements, including OpenTelemetry observability (tracing, metrics, hooks), high-speed HTTP communication via fasthttp, cryptography/ciphers, custom HTTP middleware, structured logging, request sanitization, response formatting, and error wrappers.

This toolkit serves as the shared foundation for modern microservice architectures—such as Saga-pattern e-commerce systems—by standardizing communication, logging, and metrics collection across services. It is designed from the ground up to be safe, performant, and fully compliant with security best practices like automatic PII/credential masking in payloads and logs.

---

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Practical Usage Examples](#practical-usage-examples)
  - [Observability & Structured Logging](#1-observability--structured-logging)
  - [High-Performance HTTP Client with Context Tracing](#2-high-performance-http-client-with-context-tracing)
  - [PII Masking & Log Sanitization](#3-pii-masking--log-sanitization)
  - [AES-256-CBC Encryption & Decryption](#4-aes-256-cbc-encryption--decryption)
  - [Unified API Response & Error Handling in Echo](#5-unified-api-response--error-handling-in-echo)
- [API Documentation](#api-documentation)
  - [`pkg/observability`](#pkgobservability)
  - [`pkg/app_crypto`](#pkgapp_crypto)
  - [`pkg/app_http`](#pkgapp_http)
  - [`pkg/utils`](#pkgutils)
  - [`pkg/apperror`](#pkgapperror)
  - [`pkg/response`](#pkgresponse)
  - [`pkg/pagination`](#pkgpagination)
  - [`middlewares`](#middlewares)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)
- [Maintainers & Contact](#maintainers--contact)

---

## Features

- **Telemetry & Observability**: Complete OpenTelemetry integrations supporting OTLP gRPC, OTLP HTTP, and stdout exporters. Includes standard resource builders (OS, Host, Container, Process, Service attributes).
- **High-Performance HTTP Client**: Fast HTTP client built on `valyala/fasthttp` featuring connection pooling, HTTP client-trace instrumentation, body validation, automatic JSON marshalling, multipart uploads, and structured error extraction.
- **Echo Middlewares**:
  - `RequestIDMiddleware` for generating/propagating UUID request IDs in headers and contexts.
  - `CacheWithRevalidation` for Cache-Control and validation header enforcement.
- **Logging with Trace Injection**: Zerolog integration using a custom trace hook (`TracingHook`) that automatically extracts `trace_id` and `span_id` from the context and attaches them to log entries.
- **Crypto & Cipher Suite**:
  - Advanced AES-256-CBC encryption/decryption (`Cipher`) with automated PKCS7-like padding management and secure IV initialization.
  - Classic hashing algorithms (MD5, SHA1, SHA256, SHA512, HMAC) and base64/base64URL utilities.
- **Structured PII Masking & Sanitization**:
  - Flexible string masking (Any, Left, Middle, Right) with struct tag support (`mask:"left"`).
  - High-performance recursive JSON payload and header sanitizers with customizable regex rule matching for sensitive fields (passwords, tokens, emails, credit cards, authorization headers).
- **Unified Response & Error Protocol**: Robust `AppError` interface mapping to standard HTTP status codes, combined with generic API success/failure payload builders (`SuccessResponse`, `FailedResponse`) that automatically update OpenTelemetry trace spans.

---

## Prerequisites

- **Go Version**: `1.26.1` or higher is required.
- **System Dependencies**:
  - OpenTelemetry Collector or compatible APM receiver (Jaeger, Zipkin, Datadog, Prometheus) if OTLP exporters are used.
- **Environment Configurations**:
  - No mandatory environment variables are hardcoded, but you should set standard OTel variables (e.g. `OTEL_EXPORTER_OTLP_ENDPOINT`) if utilizing OTLP export modes.

---

## Installation

Add the package to your module using `go get`:

```bash
go get github.com/dafailyasa/go-package
```

Ensure your `go.mod` is updated by running:

```bash
go mod tidy
```

---

## Practical Usage Examples

### 1. Observability & Structured Logging

Initialize the tracer and meter providers, then log messages with auto-propagated trace details.

```go
package main

import (
	"context"

	"github.com/dafailyasa/go-package/pkg/observability"
	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry Tracer
	tp, err := observability.InitTracerProvider("my-service", "1.0.0", "production", "console")
	if err != nil {
		panic(err)
	}
	defer tp.Shutdown(ctx)

	// Create Zerolog with Tracing Hook
	logger := observability.NewZeroLogHook("my-service")

	// Start a trace span
	spanCtx, span := instrumentation.NewTraceSpan(ctx, "CalculateTotals")
	defer span.End()

	// Log using the logger; span_id and trace_id are automatically appended to the log output!
	logger.Z().Info().Ctx(spanCtx).Msg("Calculating totals for user checkout")
}
```

### 2. High-Performance HTTP Client with Context Tracing

Use the wrapper around `fasthttp` to make API calls with built-in logging, request sanitization, and OpenTelemetry tracing.

```go
package main

import (
	"context"
	"fmt"

	"github.com/dafailyasa/go-package/pkg/app_http"
	"github.com/dafailyasa/go-package/pkg/observability"
)

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()
	logger := observability.NewZeroLogHook("http-client-app")

	client := app_http.NewClient(logger.Z())

	req := app_http.Request{
		Method:   "GET",
		Endpoint: "https://api.example.com/users/123",
		Headers: map[string]string{
			"Authorization": "Bearer some-sensitive-token",
		},
	}

	var user UserResponse
	resp, err := client.DoHttpRequest(ctx, req, &user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Status: %d, User ID: %s, Email: %s\n", resp.StatusCode, user.ID, user.Email)
}
```

### 3. PII Masking & Log Sanitization

Avoid leaking sensitive data in logs by masking variables, structs, or raw JSON payloads.

```go
package main

import (
	"fmt"

	"github.com/dafailyasa/go-package/pkg/utils"
)

type Profile struct {
	Password string `mask:"any"`
	Email    string `mask:"middle"`
	Phone    string `mask:"right"`
}

func main() {
	// 1. Struct Masking
	p := Profile{
		Password: "supersecretpwd",
		Email:    "johndoe@example.com",
		Phone:    "+1234567890",
	}
	utils.MaskStruct(&p)
	fmt.Printf("Masked Struct: %+v\n", p)
	// Output: Masked Struct: {Password:****** Email:***doe@ex*** Phone:+12345******}

	// 2. Raw JSON Sanitization
	rawJSON := []byte(`{"password": "secret", "email": "johndoe@example.com", "name": "John"}`)
	sanitized := utils.SanitizeBody(rawJSON)
	fmt.Println(string(sanitized))
	// Output: {"email":"***doe@ex***","name":"John","password":"******"}
}
```

### 4. AES-256-CBC Encryption & Decryption

Encrypt database columns, access keys, or internal payloads safely using `app_crypto`.

```go
package main

import (
	"fmt"

	"github.com/dafailyasa/go-package/pkg/app_crypto"
)

func main() {
	// AES key must be 32 bytes for AES-256
	key := []byte("a-very-secure-key-32-bytes-long!")
	cipher, err := app_crypto.NewES256(key)
	if err != nil {
		panic(err)
	}

	plaintext := []byte("Sensitive customer information")
	
	// Encrypt
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}

	// Decrypt
	decrypted, err := cipher.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(decrypted)) // Output: Sensitive customer information
}
```

### 5. Unified API Response & Error Handling in Echo

Wrap standard application responses and errors uniformly.

```go
package main

import (
	"errors"
	"net/http"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/dafailyasa/go-package/pkg/response"
	"github.com/labstack/echo/v4"
)

func UserHandler(c echo.Context) error {
	// Simulate user logic
	userID := c.Param("id")
	if userID == "" {
		err := apperror.BadRequest(errors.New("user ID is required"))
		return response.ErrorBuilder(err).Send(c)
	}

	data := map[string]string{"id": userID, "name": "Jane"}
	
	// Return unified success structure
	return response.SuccessBuilder(data).Send(c)
}

func main() {
	e := echo.New()
	e.GET("/users/:id", UserHandler)
	e.Start(":8080")
}
```

---

## API Documentation

### `pkg/observability`

Utilities to bootstrap and configure OpenTelemetry metrics, tracing, and structured Logging.

- **`InitTracerProvider(appName, version, environment, mode string) (*sdktrace.TracerProvider, error)`**
  Initializes global tracer. Modes: `otlp/grpc`, `otlp/http`, `console` (stdout trace).
- **`InitMeterProvider(appName, version, environment, mode string) (*sdkmetric.MeterProvider, error)`**
  Initializes global metrics recorder. Modes: `otlp/grpc`, `otlp/http`.
- **`NewZeroLogHook(appName string) *Logger`**
  Creates custom Zerolog instance configured with `TracingHook` that pulls TraceID/SpanID from contexts and appends them automatically.
- **`NewZeroLog(ctx context.Context, appName string, c ...io.Writer) *Logger`**
  Creates custom Zerolog instance using static tracing values extracted at instantiation.

---

### `pkg/app_crypto`

High-level helpers for cryptographical operations.

- **`NewCrypto(key string) *Crypto`**
  Initializes a hasher utility wrapper.
- **`(c *Crypto) EncodeSHA256(text string) string`** / **`EncodeSHA512`** / **`EncodeMD5`** / **`EncodeBASE64`** / **`DecodeBASE64`**
  Convenient hashing and encoding wrappers.
- **`(c *Crypto) EncodeSHA256HMAC(data ...string) string`** / **`EncodeSHA256HMACBase64`**
  Generates secure HMAC hashes using the configured secret key.
- **`NewES256(key []byte) (*Cipher, error)`**
  Initializes a new `Cipher` structure powered by AES-256.
- **`(c *Cipher) Encrypt(plaintext []byte) ([]byte, error)`**
  Encrypts byte slice in CBC mode with random initialization vector (IV) prepended to the ciphertext.
- **`(c *Cipher) Decrypt(ciphertext []byte) ([]byte, error)`**
  Decrypts CBC-mode ciphertext and handles padding removal.

---

### `pkg/app_http`

Fast, trace-aware HTTP client library wrapping `valyala/fasthttp`.

- **`NewClient(log *zerolog.Logger) *AppHttp`**
  Instantiates the HTTP client with robust connection pooling and timeout settings.
- **`(c *AppHttp) DoHttpRequest(ctx context.Context, req Request, res any) (*resp, error)`**
  Executes an HTTP call under a trace span. Supports JSON body (marshals non-pointers), form data, URL-encoded variables, and multipart file transfers. Automatically logs requests/responses with sensitive keys masked.

---

### `pkg/utils`

Logging sanitizers, PII masking, and HTTP header utilities.

- **`Mask(t MaskingType, s string) string`**
  Masks strings. Masking types: `any` (fully masked), `left`, `middle`, `right`.
- **`MaskStruct(s any)`**
  Iterates over struct fields and masks strings matching the `mask:"..."` struct tag.
- **`SanitizeBody(bodyBytes []byte, opts ...SanitizeOption) []byte`**
  Recursively filters JSON strings matching sensitive patterns (passwords, PII, auth).
- **`SanitizeHeaders(headers map[string]string, opts ...SanitizeOption) map[string]string`**
  Filters and masks sensitive HTTP headers (e.g. `Authorization`, `Cookie`).
- **`InjectRequestID(ctx context.Context, reqID, chainID, journeyID string) context.Context`**
  Propagates `x-request-id` into the Context.
- **`ExtractRequestID(ctx context.Context) string`**
  Retrieves request ID from Context. Generates a new UUID if missing.

---

### `pkg/apperror`

Pre-configured error registry mapping error states directly to HTTP statuses.

- **Sentinel Errors**: `ErrNotFound`, `ErrUnprocessableEntity`, `ErrForbiddenAccess`, `ErrUnauthorized`, `ErrConflict`, `ErrTimeout`, `ErrBadRequest`, `ErrGateway`.
- **`AppError`**: Error wrapper containing HTTP Code, raw Error, and Developer Message.
- **Helpers**: `BadRequest(err)`, `InternalServerError(err)`, `Unauthorized(err)`, `Forbidden(err)`, `NotFound(err)`, `Conflict(err)`, `GatewayTimeout(err)`.

---

### `pkg/response`

Standard API response formats enforcing structural uniformity across API domains.

- **`SuccessResponse`** / **`FailedResponse`**: Struct formats for standardized JSON responses.
- **`SuccessBuilder(data interface{}, pagination ...interface{}) SuccessResponse`**
  Generates success payload format. Sets span status as `Ok` when sent.
- **`ErrorBuilder(err error) FailedResponse`**
  Detects underlying type; if it is `AppError`, extracts accurate code and message. Otherwise, defaults to HTTP 500. Sets span status as `Error` when sent.

---

### `pkg/pagination`

Structured parsing and validation utilities for database query pagination.

- **`PaginationRequest`**: Parameters containing `Limit`, `Page`, `Sort`, `Status`, `Keyword`, and `Field`.
- **`(p *PaginationRequest) GetOffset() int`**
  Calculates SQL query offset (e.g., `(Page - 1) * Limit`).
- **`(p *PaginationRequest) Validate() error`**
  Fills default pagination boundaries (e.g., limit default to 10, page default to 1, sort default to `Id desc`).

---

### `middlewares`

Standard Echo middlewares to speed up routing setup.

- **`RequestIDMiddleware() echo.MiddlewareFunc`**
  Ensures every request has a request ID. Extracts or generates a UUID, appends it to HTTP headers (`X-Request-ID`), sets it in the request context, and overrides standard request structures.
- **`CacheWithRevalidation(next echo.HandlerFunc) echo.HandlerFunc`**
  Middleware enforcing cache controls: `no-cache, max-age=120, must-revalidate` along with Expires and Last-Modified headers.

---

## Troubleshooting

### "body must be a non-pointer type" when calling `DoHttpRequest`
- **Cause**: Passing a pointer to the Request structure body field, e.g., `Request{Body: &myPayload}`.
- **Solution**: Pass the payload directly by value, e.g., `Request{Body: myPayload}`. The client checks this validation on execution.

### Missing Trace ID or Span ID in logs
- **Cause**: Calling log commands without passing context, or using a context that does not contain a span created by OpenTelemetry.
- **Solution**: Ensure your operations use contexts initialized or extended via `instrumentation.NewTraceSpan(ctx, name)`. Use `.Ctx(ctx)` on Zerolog log lines.

### Cryptography "ciphertext too short" error
- **Cause**: Ciphertext being decrypted is smaller than the AES Block Size (16 bytes).
- **Solution**: Verify that database fields or payload buffers are indeed populated with output generated by `cipher.Encrypt()`.

---

## Contributing

We welcome community contributions to make this shared packages toolkit even better!

1. **Fork** the repository and create your feature branch: `git checkout -b feature/amazing-feature`.
2. **Format** your Go code before committing: `make fmt`.
3. **Verify** your changes using the linter and tests: `make vet` and `make test`. Ensure that test coverage does not drop: run `make coverage-check` to verify code coverage.
4. **Submit** your Pull Request with a clear description of the introduced changes.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.

---

## Maintainers & Contact

- **Primary Maintainer**: Dafailyasa (https://github.com/dafailyasa)
- **Repository**: [github.com/dafailyasa/go-package](https://github.com/dafailyasa/go-package)
- **Issue Tracker**: [github.com/dafailyasa/go-package/issues](https://github.com/dafailyasa/go-package/issues)
