package utils

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, output map[string]any)
	}{
		{
			name: "Positive - Mask Default Sensitive Keys",
			input: `{
				"password":"secret123",
				"token":"abcdefghijklmnop",
				"email":"john@example.com",
				"name":"John"
			}`,
			validate: func(t *testing.T, output map[string]any) {
				assert.Equal(t, DefaultMasked, output["password"])
				assert.NotEqual(t, "abcdefghijklmnop", output["token"])
				assert.NotEqual(t, "john@example.com", output["email"])
				assert.Equal(t, "John", output["name"])
			},
		},
		{
			name:  "Negative - Invalid JSON",
			input: `{invalid}`,
			validate: func(t *testing.T, output map[string]any) {
				t.Fatal("should not be called")
			},
		},
		{
			name:     "Negative - Empty Body",
			input:    "",
			validate: func(t *testing.T, output map[string]any) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SanitizeBody([]byte(tt.input))

			if tt.name == "Negative - Invalid JSON" {
				assert.Equal(t, []byte(tt.input), out)
				return
			}

			if tt.name == "Negative - Empty Body" {
				assert.Empty(t, out)
				return
			}

			var parsed map[string]any
			assert.NoError(t, json.Unmarshal(out, &parsed))
			tt.validate(t, parsed)
		})
	}
}

func TestSanitizeBodyParsed(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		out := SanitizeBodyParsed([]byte(`{"password":"123456"}`))

		result := out.(map[string]any)
		assert.Equal(t, DefaultMasked, result["password"])
	})

	t.Run("Negative - Invalid JSON", func(t *testing.T) {
		out := SanitizeBodyParsed([]byte(`invalid`))
		assert.Equal(t, "invalid", out)
	})

	t.Run("Negative - Empty Body", func(t *testing.T) {
		assert.Nil(t, SanitizeBodyParsed(nil))
	})
}

func TestSanitizeBodyIndent(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		out := SanitizeBodyIndent([]byte(`{"password":"123456"}`))

		assert.Contains(t, string(out), "\n")
		assert.Contains(t, string(out), DefaultMasked)
	})

	t.Run("Negative - Invalid JSON", func(t *testing.T) {
		out := SanitizeBodyIndent([]byte(`invalid`))
		assert.Equal(t, "invalid", string(out))
	})
}

func TestSanitizeHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		check   func(t *testing.T, result map[string]string)
	}{
		{
			name: "Positive",
			headers: map[string]string{
				"Authorization": "Bearer abcdefghijklmnop",
				"Content-Type":  "application/json",
			},
			check: func(t *testing.T, result map[string]string) {
				assert.NotEqual(t, "Bearer abcdefghijklmnop", result["Authorization"])
				assert.Equal(t, "application/json", result["Content-Type"])
			},
		},
		{
			name:    "Negative - Empty",
			headers: map[string]string{},
			check: func(t *testing.T, result map[string]string) {
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHeaders(tt.headers)
			tt.check(t, result)
		})
	}
}

func TestWithExtraKeys(t *testing.T) {
	headers := map[string]string{
		"MySecret": "abcdef",
	}

	keys := []SensitiveKey{
		{
			Pattern:     regexp.MustCompile(`(?i)^mysecret$`),
			MaskingType: Any,
		},
	}

	result := SanitizeHeaders(headers, WithExtraKeys(keys))

	assert.Equal(t, DefaultMasked, result["MySecret"])
}

func TestWithCustomKeysOnly(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer token",
		"MySecret":      "abcdef",
	}

	keys := []SensitiveKey{
		{
			Pattern:     regexp.MustCompile(`(?i)^mysecret$`),
			MaskingType: Any,
		},
	}

	result := SanitizeHeaders(headers, WithCustomKeysOnly(keys))

	// default key should no longer be masked
	assert.Equal(t, "Bearer token", result["Authorization"])

	// custom key should be masked
	assert.Equal(t, DefaultMasked, result["MySecret"])
}

func TestMatchSensitiveKey(t *testing.T) {
	keys := []SensitiveKey{
		{
			Pattern:     regexp.MustCompile(`(?i)^password$`),
			MaskingType: Any,
		},
	}

	t.Run("Positive", func(t *testing.T) {
		mask, ok := matchSensitiveKey(" password ", keys)

		assert.True(t, ok)
		assert.Equal(t, Any, mask)
	})

	t.Run("Negative", func(t *testing.T) {
		_, ok := matchSensitiveKey("username", keys)

		assert.False(t, ok)
	})
}
