package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationRequest_HasStatus(t *testing.T) {
	status := 1

	tests := []struct {
		name     string
		request  PaginationRequest
		expected bool
	}{
		{
			name: "Positive - Status Exists",
			request: PaginationRequest{
				Status: &status,
			},
			expected: true,
		},
		{
			name:     "Negative - Status Nil",
			request:  PaginationRequest{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.request.HasStatus())
		})
	}
}

func TestPaginationRequest_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected int
	}{
		{
			name:     "Positive - Custom Limit",
			limit:    25,
			expected: 25,
		},
		{
			name:     "Negative - Zero Limit Uses Default",
			limit:    0,
			expected: 10,
		},
		{
			name:     "Negative - Negative Limit",
			limit:    -5,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := PaginationRequest{
				Limit: tt.limit,
			}

			assert.Equal(t, tt.expected, req.GetLimit())
		})
	}
}

func TestPaginationRequest_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected int
	}{
		{
			name:     "Positive - Custom Page",
			page:     5,
			expected: 5,
		},
		{
			name:     "Negative - Zero Page Uses Default",
			page:     0,
			expected: 1,
		},
		{
			name:     "Negative - Negative Page",
			page:     -2,
			expected: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := PaginationRequest{
				Page: tt.page,
			}

			assert.Equal(t, tt.expected, req.GetPage())
		})
	}
}

func TestPaginationRequest_GetSort(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		expected string
	}{
		{
			name:     "Positive - Custom Sort",
			sort:     "name ASC",
			expected: "name ASC",
		},
		{
			name:     "Negative - Empty Sort Uses Default",
			sort:     "",
			expected: "Id desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := PaginationRequest{
				Sort: tt.sort,
			}

			assert.Equal(t, tt.expected, req.GetSort())
		})
	}
}

func TestPaginationRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{
			name:     "Positive - Second Page",
			page:     2,
			limit:    10,
			expected: 10,
		},
		{
			name:     "Positive - Third Page",
			page:     3,
			limit:    20,
			expected: 40,
		},
		{
			name:     "Negative - Default Values",
			page:     0,
			limit:    0,
			expected: 0,
		},
		{
			name:     "Negative - Negative Values",
			page:     -1,
			limit:    -10,
			expected: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := PaginationRequest{
				Page:  tt.page,
				Limit: tt.limit,
			}

			assert.Equal(t, tt.expected, req.GetOffset())
		})
	}
}

func TestPaginationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request PaginationRequest
	}{
		{
			name: "Positive - Custom Values",
			request: PaginationRequest{
				Page:  2,
				Limit: 20,
				Sort:  "name ASC",
			},
		},
		{
			name:    "Positive - Empty Values",
			request: PaginationRequest{},
		},
		{
			name: "Negative - Invalid Values",
			request: PaginationRequest{
				Page:  -1,
				Limit: -10,
				Sort:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			require.NoError(t, err)
		})
	}
}
