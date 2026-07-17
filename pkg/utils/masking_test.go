package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Positive - Normal String", "1234567890", DefaultMasked},
		{"Negative - Empty String", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskAny(tt.input))
		})
	}
}

func TestMaskLeft(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Positive - Long String",
			input:    "1234567890123456",
			expected: "******123456",
		},
		{
			name:     "Positive - Short String",
			input:    "1234567890",
			expected: "******7890",
		},
		{
			name:     "Negative - Empty String",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskLeft(tt.input))
		})
	}
}

func TestMaskMiddle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Positive - Long String",
			input:    "1234567890123456",
			expected: "***678901***",
		},
		{
			name:     "Positive - Short String",
			input:    "1234567890",
			expected: "***45678***",
		},
		{
			name:     "Negative - Empty String",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskMiddle(tt.input))
		})
	}
}

func TestMaskRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Positive - Long String",
			input:    "1234567890123456",
			expected: "123456******",
		},
		{
			name:     "Positive - Short String",
			input:    "1234567890",
			expected: "12345******",
		},
		{
			name:     "Negative - Empty String",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskRight(tt.input))
		})
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name     string
		maskType MaskingType
		input    string
		expected string
	}{
		{
			name:     "Mask Any",
			maskType: Any,
			input:    "secret",
			expected: DefaultMasked,
		},
		{
			name:     "Mask Left",
			maskType: Left,
			input:    "1234567890123456",
			expected: "******123456",
		},
		{
			name:     "Mask Middle",
			maskType: Middle,
			input:    "1234567890123456",
			expected: "***678901***",
		},
		{
			name:     "Mask Right",
			maskType: Right,
			input:    "1234567890123456",
			expected: "123456******",
		},
		{
			name:     "Negative - Unknown Mask Type",
			maskType: "unknown",
			input:    "secret",
			expected: "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Mask(tt.maskType, tt.input))
		})
	}
}

func TestMaskStruct(t *testing.T) {
	type User struct {
		Name     string `mask:"middle"`
		Password string `mask:"any"`
		Token    string `mask:"left"`
		Phone    string `mask:"right"`
		Address  string
	}

	tests := []struct {
		name     string
		user     User
		expected User
	}{
		{
			name: "Positive - Mask Struct",
			user: User{
				Name:     "Jonathan",
				Password: "super-secret",
				Token:    "1234567890123456",
				Phone:    "1234567890123456",
				Address:  "Jakarta",
			},
			expected: User{
				Name:     "***ath***",
				Password: DefaultMasked,
				Token:    "******123456",
				Phone:    "123456******",
				Address:  "Jakarta",
			},
		},
		{
			name: "Negative - Empty Values",
			user: User{},
			expected: User{
				Name:     "",
				Password: "",
				Token:    "",
				Phone:    "",
				Address:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.user
			MaskStruct(&user)

			assert.Equal(t, tt.expected, user)
		})
	}
}
