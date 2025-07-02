package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []int32{},
		},
		{
			name:     "empty slice",
			input:    []int32{},
			expected: []int32{},
		},
		{
			name:     "int8 slice",
			input:    []int8{1, 2, 3},
			expected: []int8{1, 2, 3},
		},
		{
			name:     "int16 slice",
			input:    []int16{1, 2, 3},
			expected: []int16{1, 2, 3},
		},
		{
			name:     "int32 slice",
			input:    []int32{1, 2, 3},
			expected: []int32{1, 2, 3},
		},
		{
			name:     "int64 slice",
			input:    []int64{1, 2, 3},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "uint slice",
			input:    []uint{1, 2, 3},
			expected: []uint{1, 2, 3},
		},
		{
			name:     "uint8 slice",
			input:    []uint8{1, 2, 3},
			expected: []uint8{1, 2, 3},
		},
		{
			name:     "uint16 slice",
			input:    []uint16{1, 2, 3},
			expected: []uint16{1, 2, 3},
		},
		{
			name:     "uint32 slice",
			input:    []uint32{1, 2, 3},
			expected: []uint32{1, 2, 3},
		},
		{
			name:     "uint64 slice",
			input:    []uint64{1, 2, 3},
			expected: []uint64{1, 2, 3},
		},
		{
			name:     "float32 slice",
			input:    []float32{1.0, 2.0, 3.0},
			expected: []float32{1.0, 2.0, 3.0},
		},
		{
			name:     "float64 slice",
			input:    []float64{1.0, 2.0, 3.0},
			expected: []float64{1.0, 2.0, 3.0},
		},
		{
			name:     "string slice to int8",
			input:    []string{"1", "2", "3"},
			expected: []int8{1, 2, 3},
		},
		{
			name:     "string slice to uint8",
			input:    []string{"1", "2", "3"},
			expected: []uint8{1, 2, 3},
		},
		{
			name:     "string slice to float32",
			input:    []string{"1.5", "2.5", "3.5"},
			expected: []float32{1.5, 2.5, 3.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case []int8:
				result := ToSlice[int8](tt.input)
				assert.Equal(t, expected, result)
			case []int16:
				result := ToSlice[int16](tt.input)
				assert.Equal(t, expected, result)
			case []int32:
				result := ToSlice[int32](tt.input)
				assert.Equal(t, expected, result)
			case []int64:
				result := ToSlice[int64](tt.input)
				assert.Equal(t, expected, result)
			case []uint:
				result := ToSlice[uint](tt.input)
				assert.Equal(t, expected, result)
			case []uint8:
				result := ToSlice[uint8](tt.input)
				assert.Equal(t, expected, result)
			case []uint16:
				result := ToSlice[uint16](tt.input)
				assert.Equal(t, expected, result)
			case []uint32:
				result := ToSlice[uint32](tt.input)
				assert.Equal(t, expected, result)
			case []uint64:
				result := ToSlice[uint64](tt.input)
				assert.Equal(t, expected, result)
			case []float32:
				result := ToSlice[float32](tt.input)
				assert.Equal(t, expected, result)
			case []float64:
				result := ToSlice[float64](tt.input)
				assert.Equal(t, expected, result)
			}
		})
	}
}

func TestToSliceE(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    any
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expected:    []int32{},
			expectError: true,
		},
		{
			name:        "empty slice",
			input:       []int32{},
			expected:    []int32{},
			expectError: false,
		},
		{
			name:        "int8 slice",
			input:       []int8{1, 2, 3},
			expected:    []int8{1, 2, 3},
			expectError: false,
		},
		{
			name:        "uint16 slice",
			input:       []uint16{1, 2, 3},
			expected:    []uint16{1, 2, 3},
			expectError: false,
		},
		{
			name:        "float32 slice",
			input:       []float32{1.0, 2.0, 3.0},
			expected:    []float32{1.0, 2.0, 3.0},
			expectError: false,
		},
		{
			name:        "invalid string slice",
			input:       []string{"1", "a", "3"},
			expected:    []int32{},
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       struct{}{},
			expected:    []int32{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case []int8:
				result, err := ToSliceE[int8](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []int16:
				result, err := ToSliceE[int16](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []int32:
				result, err := ToSliceE[int32](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []int64:
				result, err := ToSliceE[int64](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []uint:
				result, err := ToSliceE[uint](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []uint8:
				result, err := ToSliceE[uint8](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []uint16:
				result, err := ToSliceE[uint16](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []uint32:
				result, err := ToSliceE[uint32](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []uint64:
				result, err := ToSliceE[uint64](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []float32:
				result, err := ToSliceE[float32](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			case []float64:
				result, err := ToSliceE[float64](tt.input)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, expected, result)
			}
		})
	}
}

func TestToSliceDifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "int8 to int16",
			input:    []int8{1, 2, 3},
			expected: []int16{1, 2, 3},
		},
		{
			name:     "int16 to int32",
			input:    []int16{1, 2, 3},
			expected: []int32{1, 2, 3},
		},
		{
			name:     "int32 to int64",
			input:    []int32{1, 2, 3},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "uint8 to uint16",
			input:    []uint8{1, 2, 3},
			expected: []uint16{1, 2, 3},
		},
		{
			name:     "uint16 to uint32",
			input:    []uint16{1, 2, 3},
			expected: []uint32{1, 2, 3},
		},
		{
			name:     "uint32 to uint64",
			input:    []uint32{1, 2, 3},
			expected: []uint64{1, 2, 3},
		},
		{
			name:     "float32 to float64",
			input:    []float32{1.0, 2.0, 3.0},
			expected: []float64{1.0, 2.0, 3.0},
		},
		{
			name:     "int32 to float32",
			input:    []int32{1, 2, 3},
			expected: []float32{1.0, 2.0, 3.0},
		},
		{
			name:     "uint32 to float64",
			input:    []uint32{1, 2, 3},
			expected: []float64{1.0, 2.0, 3.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case []int8:
				result := ToSlice[int8](tt.input)
				assert.Equal(t, expected, result)
			case []int16:
				result := ToSlice[int16](tt.input)
				assert.Equal(t, expected, result)
			case []int32:
				result := ToSlice[int32](tt.input)
				assert.Equal(t, expected, result)
			case []int64:
				result := ToSlice[int64](tt.input)
				assert.Equal(t, expected, result)
			case []uint:
				result := ToSlice[uint](tt.input)
				assert.Equal(t, expected, result)
			case []uint8:
				result := ToSlice[uint8](tt.input)
				assert.Equal(t, expected, result)
			case []uint16:
				result := ToSlice[uint16](tt.input)
				assert.Equal(t, expected, result)
			case []uint32:
				result := ToSlice[uint32](tt.input)
				assert.Equal(t, expected, result)
			case []uint64:
				result := ToSlice[uint64](tt.input)
				assert.Equal(t, expected, result)
			case []float32:
				result := ToSlice[float32](tt.input)
				assert.Equal(t, expected, result)
			case []float64:
				result := ToSlice[float64](tt.input)
				assert.Equal(t, expected, result)
			}
		})
	}
}

func TestToAnySlice(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		var input []int
		result := ToAnySlice(input...)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("string slice", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		assert.Equal(t, []any{"a", "b", "c"}, ToAnySlice(input...))
	})

	t.Run("int slice", func(t *testing.T) {
		assert.Equal(t, []any{1, 2, 3}, ToAnySlice(1, 2, 3))
	})

	t.Run("bool slice", func(t *testing.T) {
		assert.Equal(t, []any{true, false}, ToAnySlice(true, false))
	})

	t.Run("float slice", func(t *testing.T) {
		assert.Equal(t, []any{1.1, 2.2, 3.3}, ToAnySlice(1.1, 2.2, 3.3))
	})
}
