package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Body struct {
	Length int    `db:"length"`
	Weight string `db:"weight"`
	Height int    `db:"-"`
	Age    uint
}

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"-"`
	Email string
	Body
}

func TestConvertToSliceMap(t *testing.T) {
	tests := []struct {
		data any
		want []map[string]any
	}{
		{
			data: nil,
			want: nil,
		},
		{
			data: []User{
				{ID: 1, Name: "John", Email: "john@example.com", Body: Body{Weight: "100kg", Height: 180, Age: 25}},
				{ID: 2, Name: "Jane", Email: "jane@example.com", Body: Body{Length: 1, Weight: "90kg", Height: 170, Age: 20}},
			},
			want: []map[string]any{
				{"id": 1, "weight": "100kg"},
				{"id": 2, "length": 1, "weight": "90kg"},
			},
		},
		{
			data: []*User{
				{ID: 1, Name: "John", Email: "john@example.com", Body: Body{Weight: "100kg", Height: 180, Age: 25}},
				{ID: 2, Name: "Jane", Email: "jane@example.com", Body: Body{Length: 1, Weight: "90kg", Height: 170, Age: 20}},
			},
			want: []map[string]any{
				{"id": 1, "weight": "100kg"},
				{"id": 2, "length": 1, "weight": "90kg"},
			},
		},
		{
			data: []Body{
				{Weight: "100kg", Height: 180, Age: 25},
				{Length: 1, Weight: "90kg", Height: 170, Age: 20},
			},
			want: []map[string]any{{"weight": "100kg"}, {"length": 1, "weight": "90kg"}},
		},
		{
			data: Body{
				Weight: "100kg",
				Height: 180,
				Age:    25,
			},
			want: []map[string]any{{"weight": "100kg"}},
		},
		{
			data: &Body{
				Weight: "100kg",
				Height: 180,
				Age:    25,
			},
			want: []map[string]any{{"weight": "100kg"}},
		},
		{
			data: map[string]any{
				"weight": "100kg",
				"Age":    25,
			},
			want: []map[string]any{{"weight": "100kg", "Age": 25}},
		},
		{
			data: []map[string]any{
				{"weight": "100kg", "Age": 25},
				{"weight": "90kg", "Age": 20},
			},
			want: []map[string]any{{"weight": "100kg", "Age": 25}, {"weight": "90kg", "Age": 20}},
		},
	}

	for _, test := range tests {
		sliceMap, err := convertToSliceMap(test.data)
		assert.NoError(t, err)
		assert.Equal(t, test.want, sliceMap)
	}
}
