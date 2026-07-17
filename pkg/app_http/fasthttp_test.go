package app_http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dafailyasa/go-package/pkg/app_http"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AppHTTPSuite struct {
	suite.Suite

	logger zerolog.Logger
	client *app_http.AppHttp
}

func TestAppHTTPSuite(t *testing.T) {
	suite.Run(t, new(AppHTTPSuite))
}

func (s *AppHTTPSuite) SetupTest() {
	s.logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	s.client = app_http.NewClient(&s.logger)
}

func (s *AppHTTPSuite) TestDoHttpRequestError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   http.MethodGet,
		Endpoint: server.URL + "/nonexistent",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	var res struct {
		Error string `json:"error"`
	}

	resp, err := s.client.DoHttpRequest(context.Background(), req, &res)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
	assert.Equal(s.T(), "not found", res.Error)
}

func (s *AppHTTPSuite) TestDoHttpRequestWithJSONBody() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "value", body["key"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":      true,
			"access_token": "123456",
		})
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   http.MethodPost,
		Endpoint: server.URL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]string{
			"key": "value",
		},
	}

	var res struct {
		Success     bool   `json:"success"`
		AccessToken string `json:"access_token"`
	}

	_, err := s.client.DoHttpRequest(context.Background(), req, &res)

	require.NoError(s.T(), err)
	assert.True(s.T(), res.Success)
	assert.Equal(s.T(), "123456", res.AccessToken)
}

func (s *AppHTTPSuite) TestDoHttpRequestWithFormFile() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(s.T(), err)

		file, _, err := r.FormFile("file")
		require.NoError(s.T(), err)
		defer file.Close()

		json.NewEncoder(w).Encode(map[string]bool{
			"success": true,
		})
	}))
	defer server.Close()

	tmp, err := os.CreateTemp("", "*.txt")
	require.NoError(s.T(), err)

	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = tmp.WriteString("hello world")
	require.NoError(s.T(), err)

	req := app_http.Request{
		Method:   http.MethodPost,
		Endpoint: server.URL,
		Headers: map[string]string{
			"Content-Type": "multipart/form-data",
		},
		Files: map[string]app_http.File{
			"file": {
				FileName: "test.txt",
				File:     tmp,
			},
		},
	}

	var res struct {
		Success bool `json:"success"`
	}

	_, err = s.client.DoHttpRequest(context.Background(), req, &res)

	require.NoError(s.T(), err)
	assert.True(s.T(), res.Success)
}

func (s *AppHTTPSuite) TestDoHttpRequestWithURLEncodedBody() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(
			s.T(),
			"application/x-www-form-urlencoded",
			r.Header.Get("Content-Type"),
		)

		require.NoError(s.T(), r.ParseForm())

		assert.Equal(s.T(), "john", r.FormValue("username"))
		assert.Equal(s.T(), "secret", r.FormValue("password"))

		json.NewEncoder(w).Encode(map[string]bool{
			"success": true,
		})
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   http.MethodPost,
		Endpoint: server.URL,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		FormEncoded: map[string]string{
			"username": "john",
			"password": "secret",
		},
	}

	var res struct {
		Success bool `json:"success"`
	}

	_, err := s.client.DoHttpRequest(context.Background(), req, &res)

	require.NoError(s.T(), err)
	assert.True(s.T(), res.Success)
}
