package mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertMapHeadersToSlice(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    []string
	}{
		{
			name:    "empty map",
			headers: map[string]string{},
			want:    []string{},
		},
		{
			name: "single header",
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			want: []string{"Content-Type: text/plain"},
		},
		{
			name: "multiple headers",
			headers: map[string]string{
				"Content-Type": "text/plain",
				"From":         "test@example.com",
				"To":           "recipient@example.com",
			},
			want: []string{
				"Content-Type: text/plain",
				"From: test@example.com",
				"To: recipient@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMapHeadersToSlice(tt.headers)
			assert.ElementsMatch(t, tt.want, got, "convertMapHeadersToSlice() = %v, want %v", got, tt.want)
		})
	}
}

func TestConvertSliceHeadersToMap(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    map[string]string
	}{
		{
			name:    "empty slice",
			headers: []string{},
			want:    map[string]string{},
		},
		{
			name:    "single header",
			headers: []string{"Content-Type: text/plain"},
			want: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name: "multiple headers",
			headers: []string{
				"Content-Type: text/plain",
				"From: test@example.com",
				"To: recipient@example.com",
			},
			want: map[string]string{
				"Content-Type": "text/plain",
				"From":         "test@example.com",
				"To":           "recipient@example.com",
			},
		},
		{
			name: "headers with extra spaces",
			headers: []string{
				"Content-Type : text/plain",
				"From : test@example.com",
				"To : recipient@example.com",
			},
			want: map[string]string{
				"Content-Type": "text/plain",
				"From":         "test@example.com",
				"To":           "recipient@example.com",
			},
		},
		{
			name: "invalid header format",
			headers: []string{
				"Content-Type: text/plain",
				"InvalidHeader",
				"To: recipient@example.com",
			},
			want: map[string]string{
				"Content-Type": "text/plain",
				"To":           "recipient@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSliceHeadersToMap(tt.headers)
			assert.Equal(t, tt.want, got, "convertSliceHeadersToMap() = %v, want %v", got, tt.want)
		})
	}
}
