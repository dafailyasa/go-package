package response

import (
	"net/http"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/dafailyasa/go-package/pkg/observability/instrumentation"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	INTERNAL_SERVER_ERROR = "internal_server_error"
	SUCCESS               = "success"
)

// FailedResponse represents a failed response structure for API responses.
type FailedResponse struct {
	Code    int    `json:"code" example:"500"`                      // HTTP status code.
	Message string `json:"message" example:"internal_server_error"` // Message corresponding to the status code.
	Error   string `json:"error" example:"{$err}"`                  // error message.
}

// ErrorBuilder constructs a FailedResponse based on the provided error.
func ErrorBuilder(err error) FailedResponse {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		ae := err.(*apperror.AppError)

		return FailedResponse{
			Code:    ae.Code,
			Message: ae.Message,
			Error:   ae.Error(),
		}
	}

	var errString = INTERNAL_SERVER_ERROR
	if err != nil {
		errString = err.Error()
	}

	return FailedResponse{
		Code:    http.StatusInternalServerError,
		Message: INTERNAL_SERVER_ERROR,
		Error:   errString,
	}
}

// Send sends the CustomResponse as a JSON response using the provided Echo context.
func (x FailedResponse) Send(c echo.Context) error {
	instrumentation.RecordSpanError(trace.SpanFromContext(c.Request().Context()), errors.New(x.Error))

	return c.JSON(x.Code, x)
}

// SuccessResponse represents a success response structure for API responses.
type SuccessResponse struct {
	Success
	Pagination
}

type ResponseFormat struct {
	Code    int    `json:"code" example:"200"` // HTTP status code.
	Message string `json:"message" example:"success"`
}

type Success struct {
	ResponseFormat
	Data interface{} `json:"data,omitempty"` // data payload.
}

type Pagination struct {
	Pagination interface{} `json:"pagination,omitempty"` //pagination payload.
	Success
}

// SuccessBuilder constructs a CustomResponse with a Success status and the provided response data.
func SuccessBuilder(response interface{}, pagination ...interface{}) SuccessResponse {
	result := SuccessResponse{
		Success: Success{
			ResponseFormat: ResponseFormat{
				Code:    http.StatusOK,
				Message: SUCCESS,
			},
			Data: response,
		},
	}

	if len(pagination) > 0 {
		result.Pagination.Pagination = pagination[0]
	}

	return result
}

// Send sends the CustomResponse as a JSON response using the provided Echo context.
func (c SuccessResponse) Send(ctx echo.Context) error {
	trace.SpanFromContext(ctx.Request().Context()).SetStatus(codes.Ok, http.StatusText(c.Code))
	return ctx.JSON(c.Code, c)
}
