package utils

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SanitizerSuite struct {
	suite.Suite
}

func TestSanitizerSuite(t *testing.T) {
	suite.Run(t, new(SanitizerSuite))
}

func (s *SanitizerSuite) TestSanitizeBody() {
	tests := []struct {
		name     string
		input    string
		validate func(map[string]any)
	}{
		{
			name: "Positive - Mask Default Sensitive Keys",
			input: `{
				"password":"secret123",
				"token":"abcdefghijklmnop",
				"email":"john@example.com",
				"name":"John"
			}`,
			validate: func(output map[string]any) {
				s.Equal(DefaultMasked, output["password"])
				s.NotEqual("abcdefghijklmnop", output["token"])
				s.NotEqual("john@example.com", output["email"])
				s.Equal("John", output["name"])
			},
		},
		{
			name:  "Negative - Invalid JSON",
			input: `{invalid}`,
		},
		{
			name:  "Negative - Empty Body",
			input: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			out := SanitizeBody([]byte(tt.input))

			switch tt.name {
			case "Negative - Invalid JSON":
				s.Equal([]byte(tt.input), out)

			case "Negative - Empty Body":
				s.Empty(out)

			default:
				var parsed map[string]any
				s.NoError(json.Unmarshal(out, &parsed))
				tt.validate(parsed)
			}
		})
	}
}

func (s *SanitizerSuite) TestSanitizeBodyParsed() {
	s.Run("Positive", func() {
		out := SanitizeBodyParsed([]byte(`{"password":"123456"}`))

		result, ok := out.(map[string]any)
		s.True(ok)
		s.Equal(DefaultMasked, result["password"])
	})

	s.Run("Negative - Invalid JSON", func() {
		out := SanitizeBodyParsed([]byte(`invalid`))
		s.Equal("invalid", out)
	})

	s.Run("Negative - Empty Body", func() {
		s.Nil(SanitizeBodyParsed(nil))
	})
}

func (s *SanitizerSuite) TestSanitizeBodyIndent() {
	s.Run("Positive", func() {
		out := SanitizeBodyIndent([]byte(`{"password":"123456"}`))

		s.Contains(string(out), "\n")
		s.Contains(string(out), DefaultMasked)
	})

	s.Run("Negative - Invalid JSON", func() {
		out := SanitizeBodyIndent([]byte(`invalid`))
		s.Equal("invalid", string(out))
	})
}

func (s *SanitizerSuite) TestSanitizeHeaders() {
	tests := []struct {
		name    string
		headers map[string]string
		check   func(map[string]string)
	}{
		{
			name: "Positive",
			headers: map[string]string{
				"Authorization": "Bearer abcdefghijklmnop",
				"Content-Type":  "application/json",
			},
			check: func(result map[string]string) {
				s.NotEqual("Bearer abcdefghijklmnop", result["Authorization"])
				s.Equal("application/json", result["Content-Type"])
			},
		},
		{
			name:    "Negative - Empty",
			headers: map[string]string{},
			check: func(result map[string]string) {
				s.Empty(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := SanitizeHeaders(tt.headers)
			tt.check(result)
		})
	}
}

func (s *SanitizerSuite) TestWithExtraKeys() {
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

	s.Equal(DefaultMasked, result["MySecret"])
}

func (s *SanitizerSuite) TestWithCustomKeysOnly() {
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

	s.Equal("Bearer token", result["Authorization"])
	s.Equal(DefaultMasked, result["MySecret"])
}

func (s *SanitizerSuite) TestMatchSensitiveKey() {
	keys := []SensitiveKey{
		{
			Pattern:     regexp.MustCompile(`(?i)^password$`),
			MaskingType: Any,
		},
	}

	s.Run("Positive", func() {
		mask, ok := matchSensitiveKey(" password ", keys)

		s.True(ok)
		s.Equal(Any, mask)
	})

	s.Run("Negative", func() {
		_, ok := matchSensitiveKey("username", keys)

		s.False(ok)
	})
}
