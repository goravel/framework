package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/carbon"
)

type Body struct {
	Length   int     `db:"length"`
	Weight   string  `db:"weight"`
	Head     *string `db:"head"`
	Height   int     `db:"-"`
	Age      uint    `db:"-"`
	DateTime carbon.DateTime
	leg      int `db:"leg"`
}

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"-"`
	Email string
	Body
	TestSoftDeletes
	TestTimestamps
}

type TestSoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" db:"deleted_at"`
}

type TestTimestamps struct {
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" db:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" db:"updated_at"`
}

func TestConvertToSliceMap(t *testing.T) {
	testNow := time.Now()
	deletedAt := gorm.DeletedAt{Time: testNow, Valid: true}
	dateTime := carbon.NewDateTime(carbon.FromStdTime(testNow))
	head := "head"

	tests := []struct {
		name string
		data any
		want []map[string]any
	}{
		{
			name: "nil",
			data: nil,
			want: nil,
		},
		{
			name: "slice",
			data: []User{
				{ID: 1, Name: "John", Email: "john@example.com", Body: Body{Weight: "100kg", Height: 180, Head: &head, Age: 25, DateTime: *dateTime, leg: 1},
					TestSoftDeletes: TestSoftDeletes{DeletedAt: deletedAt},
					TestTimestamps:  TestTimestamps{CreatedAt: dateTime, UpdatedAt: dateTime}},
				{ID: 2, Name: "Jane", Email: "jane@example.com", Body: Body{Length: 1, Weight: "90kg", Height: 170, Head: &head, Age: 20, DateTime: *dateTime, leg: 2},
					TestSoftDeletes: TestSoftDeletes{DeletedAt: deletedAt},
					TestTimestamps:  TestTimestamps{CreatedAt: dateTime, UpdatedAt: dateTime}},
			},
			want: []map[string]any{
				{"id": 1, "email": "john@example.com", "weight": "100kg", "head": &head, "date_time": *dateTime, "created_at": dateTime, "updated_at": dateTime, "deleted_at": deletedAt},
				{"id": 2, "email": "jane@example.com", "length": 1, "weight": "90kg", "head": &head, "date_time": *dateTime, "created_at": dateTime, "updated_at": dateTime, "deleted_at": deletedAt},
			},
		},
		{
			name: "slice with pointer",
			data: []*User{
				{ID: 1, Name: "John", Email: "john@example.com", Body: Body{Weight: "100kg", Height: 180, Head: &head, Age: 25, DateTime: *dateTime, leg: 1},
					TestSoftDeletes: TestSoftDeletes{DeletedAt: deletedAt},
					TestTimestamps:  TestTimestamps{CreatedAt: dateTime, UpdatedAt: dateTime}},
				{ID: 2, Name: "Jane", Email: "jane@example.com", Body: Body{Length: 1, Weight: "90kg", Height: 170, Head: &head, Age: 20, DateTime: *dateTime, leg: 2},
					TestSoftDeletes: TestSoftDeletes{DeletedAt: deletedAt},
					TestTimestamps:  TestTimestamps{CreatedAt: dateTime, UpdatedAt: dateTime}},
			},
			want: []map[string]any{
				{"id": 1, "email": "john@example.com", "weight": "100kg", "head": &head, "date_time": *dateTime, "created_at": dateTime, "updated_at": dateTime, "deleted_at": deletedAt},
				{"id": 2, "email": "jane@example.com", "length": 1, "weight": "90kg", "head": &head, "date_time": *dateTime, "created_at": dateTime, "updated_at": dateTime, "deleted_at": deletedAt},
			},
		},
		{
			name: "struct",
			data: Body{
				Weight:   "100kg",
				Height:   180,
				Head:     &head,
				Age:      25,
				DateTime: *dateTime,
				leg:      1,
			},
			want: []map[string]any{{"weight": "100kg", "head": &head, "date_time": *dateTime}},
		},
		{
			name: "pointer",
			data: &Body{
				Weight:   "100kg",
				Height:   180,
				Head:     &head,
				Age:      25,
				DateTime: *dateTime,
				leg:      1,
			},
			want: []map[string]any{{"weight": "100kg", "head": &head, "date_time": *dateTime}},
		},
		{
			name: "map",
			data: map[string]any{
				"weight": "100kg",
				"Age":    25,
			},
			want: []map[string]any{{"weight": "100kg", "Age": 25}},
		},
		{
			name: "slice of map",
			data: []map[string]any{
				{"weight": "100kg", "Age": 25},
				{"weight": "90kg", "Age": 20},
			},
			want: []map[string]any{{"weight": "100kg", "Age": 25}, {"weight": "90kg", "Age": 20}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sliceMap, err := convertToSliceMap(tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, sliceMap)
		})
	}
}
