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
)

func TestDoHttpRequestError(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	appHttp := app_http.NewClient(&logger)

	// Create a test server that returns an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a 404 Not Found error
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   "GET",
		Endpoint: server.URL + "/nonexistent", // Use the test server URL
		Headers:  map[string]string{"Content-Type": "application/json"},
	}

	var res struct {
		Error string `json:"error"`
	}

	// Call the DoHttpRequest method
	resp, err := appHttp.DoHttpRequest(context.Background(), req, &res)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
	assert.Contains(t, res.Error, "not found")
}

func TestDoHttpRequestWithJSONBody(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	appHttp := app_http.NewClient(&logger)

	// Create a test server that returns a mocked response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Respond with the expected JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"success": true, "access_token": "123456"})
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   "POST",
		Endpoint: server.URL + "/test",
		Headers:  map[string]string{"Content-Type": "application/json"},
		Body:     map[string]string{"key": "value"},
	}

	var res struct {
		Success     bool   `json:"success"`
		AccessToken string `json:"access_token"`
	}

	// Call the DoHttpRequest method
	_, err := appHttp.DoHttpRequest(context.Background(), req, &res)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func TestDoHttpRequestWithFormFile(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	appHttp := app_http.NewClient(&logger)

	// Create a test server that handles form file uploads
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form
		err := r.ParseMultipartForm(10 << 20) // Limit your max memory
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		// Check if the file exists
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File not found", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Respond with a success message
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up

	// Write some test data to the file
	if _, err := tempFile.WriteString("This is a test file."); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file so it can be read later
	defer tempFile.Close()

	// Prepare the request with the file
	req := app_http.Request{
		Method:   http.MethodPost,
		Endpoint: server.URL + "/upload",
		Headers:  map[string]string{"Content-Type": "multipart/form-data"},
		Files: map[string]app_http.File{
			"file": {
				FileName: "fileku.text",
				File:     tempFile,
			},
		},
	}

	var res struct {
		Success bool `json:"success"`
	}

	// Call the DoHttpRequest method
	_, err = appHttp.DoHttpRequest(context.Background(), req, &res)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func TestDoHttpRequestWithURLEncodedBody(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	appHttp := app_http.NewClient(&logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t,
			"application/x-www-form-urlencoded",
			r.Header.Get("Content-Type"),
		)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, "john", r.FormValue("username"))
		assert.Equal(t, "secret", r.FormValue("password"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"success": true,
		})
	}))
	defer server.Close()

	req := app_http.Request{
		Method:   http.MethodPost,
		Endpoint: server.URL + "/login",
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

	_, err := appHttp.DoHttpRequest(context.Background(), req, &res)

	assert.NoError(t, err)
	assert.True(t, res.Success)
}
