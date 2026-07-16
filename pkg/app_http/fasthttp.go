package app_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
	"github.com/dafailyasa/go-package/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const ()

type (
	File struct {
		FileName string
		File     io.Reader
	}

	Request struct {
		Method      string
		Endpoint    string
		Headers     map[string]string
		Body        any
		Files       map[string]File
		FormData    map[string]string
		FormEncoded map[string]string
	}
)

type resp struct {
	StatusCode int
	Headers    map[string]string
}

// validateBody checks if the provided body is a non-pointer type
func (r Request) validateBody() error {
	if r.Body == nil {
		return nil
	}

	v := reflect.ValueOf(r.Body)

	// Check if the body is a pointer
	if v.Kind() == reflect.Ptr {
		return apperror.ErrBodyMustNonPointerType
	}

	return nil
}

type AppHttp struct {
	client *fasthttp.Client
	log    *zerolog.Logger
}

// NewClient creates a new fasthttp client with default settings
func NewClient(log *zerolog.Logger) *AppHttp {
	return &AppHttp{
		client: &fasthttp.Client{
			// Maximum number of connections allowed per host. This controls the number of keep-alive connections.
			MaxConnsPerHost: 50,

			// The function used to establish network connections. The default is sufficient for most cases.
			Dial: fasthttp.Dial,

			// The maximum time a connection can remain idle before being closed.
			MaxIdleConnDuration: 30 * time.Second,

			// Maximum time allowed for reading a response from the server.
			ReadTimeout: 30 * time.Second,

			// Maximum time allowed for writing a request to the server.
			WriteTimeout: 30 * time.Second,
		},
		log: log,
	}
}

func (c *AppHttp) DoHttpRequest(ctx context.Context, req Request, res any) (*resp, error) {
	ctx, span := instrumentation.NewTraceSpan(ctx, "DoHttpRequest")
	defer span.End()
	if err := req.validateBody(); err != nil {
		return nil, err
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.Header.DisableNormalizing()
	request.Header.SetMethod(req.Method)
	request.SetRequestURI(req.Endpoint)

	// Set request headers
	for key, value := range req.Headers {
		request.Header.Set(key, value)
	}

	// If there are files to upload, create multipart form data
	if req.FormEncoded != nil {
		values := url.Values{}

		for k, v := range req.FormEncoded {
			values.Set(k, v)
		}

		request.SetBodyString(values.Encode())
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if req.Files != nil || req.FormData != nil {
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)

		// Add files to the form
		for key, file := range req.Files {
			part, err := writer.CreateFormFile(key, file.FileName)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create form file")
			}

			if _, err := io.Copy(part, file.File); err != nil {
				return nil, errors.Wrap(err, "failed to copy file to form")
			}
		}

		//Add param to form
		for key, param := range req.FormData {
			_ = writer.WriteField(key, param)
		}

		// Close the writer to finalize the form data
		if err := writer.Close(); err != nil {
			return nil, errors.Wrap(err, "failed to close writer")
		}

		request.SetBody(buffer.Bytes())
		request.Header.Set("Content-Type", writer.FormDataContentType())
	} else if req.Body != nil {
		// If there is a body, marshal it to JSON
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			c.log.Err(err).Ctx(ctx).Msg("[DoHttpRequest]Marshal")

			return nil, errors.Wrap(err, "failed to marshal request body")
		}

		request.SetBody(jsonBody)
		request.Header.Set("Content-Type", "application/json")
	}

	requestID := utils.ExtractRequestID(ctx)

	start := time.Now()
	c.log.Info().Ctx(ctx).
		Str("x-request-id", requestID).
		Str("method", req.Method).
		Str("endpoint", req.Endpoint).
		Interface("headers", req.Headers).
		Interface("requestBody", c.sanitizeRequestBody(req.Body)).
		Interface("requestForm", c.sanitizeRequestBody(req.FormData)).
		Interface("requestFormEncoded", c.sanitizeRequestBody(req.FormEncoded)).
		Interface("requestFiles", req.Files).
		Msg("[DoHttpRequest]request")

	// Perform the request
	if err := c.client.Do(request, response); err != nil {
		c.log.Err(err).Ctx(ctx).
			Str("method", req.Method).
			Str("endpoint", req.Endpoint).
			Msg("[DoHttpRequest]client.Do")

		return nil, errors.Wrap(err, "failed to execute HTTP request")
	}

	if response == nil {
		c.log.Info().Ctx(ctx).
			Str("x-request-id", requestID).
			Str("method", req.Method).
			Str("endpoint", req.Endpoint).
			Dur("duration", time.Since(start)).
			Msg("[DoHttpRequest]UnexpectedError: No response received")

		return nil, fmt.Errorf("UnexpectedError: No response received from %v", req.Endpoint)
	}

	if strings.TrimSpace(string(response.Body())) != "" {
		if err := json.Unmarshal(response.Body(), res); err != nil {
			c.log.Err(err).Ctx(ctx).
				Str("x-request-id", requestID).
				Str("method", req.Method).
				Str("endpoint", req.Endpoint).
				Interface("requestBody", c.sanitizeRequestBody(string(response.Body()))).
				Interface("responseBody", res).
				Msg("[DoHttpRequest]json.Unmarshal")

			return nil, errors.Wrap(err, "failed to decode response")
		}
	}

	c.log.Info().Ctx(ctx).
		Str("x-request-id", requestID).
		Str("method", req.Method).
		Str("endpoint", req.Endpoint).
		Int("statusCode", response.StatusCode()).
		Interface("responseBody", res).
		Dur("duration", time.Since(start)).Msg("[DoHttpRequest]response")

	headers := make(map[string]string)

	for key, value := range response.Header.All() {
		headers[string(key)] = string(value)
	}

	resp := resp{
		StatusCode: response.StatusCode(),
		Headers:    headers,
	}

	return &resp, nil
}
func (c *AppHttp) sanitizeRequestBody(body any) any {
	if body == nil {
		return nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "[unserializable request body]"
	}
	return utils.SanitizeBody(raw)
}
