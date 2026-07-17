package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorBuilder(t *testing.T) {
	tests := []struct {
		name            string
		input           error
		expectedCode    int
		expectedMessage string
		expectedError   string
	}{
		{
			name: "Positive - AppError",
			input: apperror.BadRequest(
				errors.New("invalid payload"),
			),
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "bad_request",
			expectedError:   "invalid payload",
		},
		{
			name:            "Negative - Generic Error",
			input:           errors.New("database error"),
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: INTERNAL_SERVER_ERROR,
			expectedError:   "database error",
		},
		{
			name:            "Negative - Nil Error",
			input:           nil,
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: INTERNAL_SERVER_ERROR,
			expectedError:   INTERNAL_SERVER_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ErrorBuilder(tt.input)

			assert.Equal(t, tt.expectedCode, resp.Code)
			assert.Equal(t, tt.expectedMessage, resp.Message)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}

func TestSuccessResponse_Send(t *testing.T) {
	tests := []struct {
		name string
		resp SuccessResponse
	}{
		{
			name: "Positive - Success Response",
			resp: SuccessBuilder(map[string]string{
				"hello": "world",
			}),
		},
		{
			name: "Positive - Nil Data",
			resp: SuccessBuilder(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := tt.resp.Send(ctx)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		})
	}
}

func TestFailedResponse_Send(t *testing.T) {
	tests := []struct {
		name string
		resp FailedResponse
	}{
		{
			name: "Positive - Bad Request",
			resp: FailedResponse{
				Code:    http.StatusBadRequest,
				Message: "bad_request",
				Error:   "invalid payload",
			},
		},
		{
			name: "Positive - Internal Server Error",
			resp: FailedResponse{
				Code:    http.StatusInternalServerError,
				Message: INTERNAL_SERVER_ERROR,
				Error:   "unexpected error",
			},
		},
		{
			name: "Negative - Empty Error Message",
			resp: FailedResponse{
				Code:    http.StatusInternalServerError,
				Message: INTERNAL_SERVER_ERROR,
				Error:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := tt.resp.Send(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.resp.Code, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		})
	}
}
