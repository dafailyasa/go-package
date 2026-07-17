package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dafailyasa/go-package/pkg/apperror"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type ResponseSuite struct {
	suite.Suite

	echo *echo.Echo
}

func TestResponseSuite(t *testing.T) {
	suite.Run(t, new(ResponseSuite))
}

func (s *ResponseSuite) SetupTest() {
	s.echo = echo.New()
}

func (s *ResponseSuite) TestErrorBuilder() {
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
		s.Run(tt.name, func() {
			resp := ErrorBuilder(tt.input)

			s.Equal(tt.expectedCode, resp.Code)
			s.Equal(tt.expectedMessage, resp.Message)
			s.Equal(tt.expectedError, resp.Error)
		})
	}
}

func (s *ResponseSuite) TestSuccessResponseSend() {
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
		{
			name: "Positive - With Pagination",
			resp: SuccessBuilder(
				[]string{"a", "b"},
				map[string]any{
					"page":  1,
					"limit": 10,
					"total": 2,
				},
			),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := s.echo.NewContext(req, rec)

			err := tt.resp.Send(ctx)

			s.NoError(err)
			s.Equal(http.StatusOK, rec.Code)
			s.NotEmpty(rec.Body.String())
		})
	}
}

func (s *ResponseSuite) TestFailedResponseSend() {
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
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := s.echo.NewContext(req, rec)

			err := tt.resp.Send(ctx)

			s.NoError(err)
			s.Equal(tt.resp.Code, rec.Code)
			s.NotEmpty(rec.Body.String())
		})
	}
}
