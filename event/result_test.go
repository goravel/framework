package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestResult(t *testing.T) {
	tests := []struct {
		name     string
		errs     []error
		failed   bool
		expected string
	}{
		{name: "Empty"},
		{name: "Nil", errs: nil},
		{name: "Single", errs: []error{errors.New("first")}, failed: true, expected: "first"},
		{
			name:     "Multiple",
			errs:     []error{errors.New("first"), errors.New("second")},
			failed:   true,
			expected: "first\nsecond",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := NewResult(test.errs)

			assert.Equal(t, test.failed, result.Failed())
			assert.Equal(t, test.errs, result.Errors())

			if test.expected == "" {
				assert.NoError(t, result.Error())
				return
			}

			assert.EqualError(t, result.Error(), test.expected)
		})
	}
}
