package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/carbon"
)

type Json struct {
	A string
	B int
	C []string
	D []int
}

func (r *Json) Value() (driver.Value, error) {
	bytes, err := json.Marshal(r)
	return string(bytes), err
}

func (r *Json) Scan(value any) (err error) {
	if data, ok := value.([]byte); ok && len(data) > 0 {
		err = json.Unmarshal(data, &r)
	}
	return
}

func TestNewRow(t *testing.T) {
	tests := []struct {
		name string
		row  map[string]any
		err  error
	}{
		{
			name: "create row with nil error",
			row:  map[string]any{"id": 1, "name": "test"},
			err:  nil,
		},
		{
			name: "create row with error",
			row:  nil,
			err:  errors.New("test error"),
		},
		{
			name: "create row with empty map",
			row:  map[string]any{},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.row, tt.err)
			assert.NotNil(t, row)
			assert.Equal(t, tt.err, row.err)
			assert.Equal(t, tt.row, row.row)
		})
	}
}

func TestRow_Err(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "nil error",
			err:  nil,
		},
		{
			name: "with error",
			err:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(nil, tt.err)
			assert.Equal(t, tt.err, row.Err())
		})
	}
}

func TestScan_Basics(t *testing.T) {
	type CustomInt int
	type CustomString string

	type TestStruct struct {
		ID         int
		Name       string
		AgentId    string
		UserID     int
		UserName   string
		Slice      []string
		Map        map[string]any
		Json       Json
		Active     bool
		Score      float64
		Count      int64
		Total      uint
		Amount     float32
		Age        int32
		Level      uint32
		Size       int16
		Priority   uint16
		Flag       int8
		Status     uint8
		CustomInt  CustomInt
		CustomStr  CustomString
		NamePtr    *string
		AgePtr     *int
		ActivePtr  *bool
		ScorePtr   *float64
		unexported string
	}

	tests := []struct {
		name      string
		rowData   map[string]any
		rowErr    error
		target    any
		wantErr   bool
		assertion func(t *testing.T, target any)
	}{
		{
			name:    "scan with existing error",
			rowData: map[string]any{"id": 1},
			rowErr:  assert.AnError,
			target:  &TestStruct{},
			wantErr: true,
		},
		{
			name:    "scan with nil target",
			rowData: map[string]any{"id": 1},
			rowErr:  nil,
			target:  nil,
			wantErr: true,
		},
		{
			name:    "scan with non-pointer target",
			rowData: map[string]any{"id": 1},
			rowErr:  nil,
			target:  TestStruct{},
			wantErr: true,
		},
		{
			name: "scan struct with basic types",
			rowData: map[string]any{
				"id":        1,
				"name":      "test",
				"user_id":   10,
				"user_name": "john",
				"agent_id":  "agent",
				"slice":     `["a", "b"]`,
				"map":       `{"a": "b", "c": 1}`,
				"json":      `{"a": "a", "b": 2, "c": ["x", "y"], "d": [3,4]}`,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, "test", result.Name)
				assert.Equal(t, 10, result.UserID)
				assert.Equal(t, "john", result.UserName)
				assert.Equal(t, "agent", result.AgentId)
				assert.Equal(t, []string{"a", "b"}, result.Slice)
				assert.Equal(t, map[string]any{"a": "b", "c": float64(1)}, result.Map)
				assert.Equal(t, Json{A: "a", B: 2, C: []string{"x", "y"}, D: []int{3, 4}}, result.Json)
			},
		},
		{
			name: "scan struct with all numeric types",
			rowData: map[string]any{
				"id":       1,
				"count":    int64(100),
				"total":    uint(200),
				"score":    98.5,
				"amount":   float32(45.5),
				"age":      int32(30),
				"level":    uint32(5),
				"size":     int16(10),
				"priority": uint16(3),
				"flag":     int8(1),
				"status":   uint8(2),
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, int64(100), result.Count)
				assert.Equal(t, uint(200), result.Total)
				assert.Equal(t, 98.5, result.Score)
				assert.Equal(t, float32(45.5), result.Amount)
				assert.Equal(t, int32(30), result.Age)
				assert.Equal(t, uint32(5), result.Level)
				assert.Equal(t, int16(10), result.Size)
				assert.Equal(t, uint16(3), result.Priority)
				assert.Equal(t, int8(1), result.Flag)
				assert.Equal(t, uint8(2), result.Status)
			},
		},
		{
			name: "scan struct with boolean",
			rowData: map[string]any{
				"id":     1,
				"active": true,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.True(t, result.Active)
			},
		},
		{
			name: "scan struct with custom types",
			rowData: map[string]any{
				"id":         1,
				"custom_int": 42,
				"custom_str": "custom",
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, CustomInt(42), result.CustomInt)
				assert.Equal(t, CustomString("custom"), result.CustomStr)
			},
		},
		{
			name: "scan struct with pointer fields",
			rowData: map[string]any{
				"id":         1,
				"name_ptr":   "pointer",
				"age_ptr":    25,
				"active_ptr": true,
				"score_ptr":  88.5,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.NotNil(t, result.NamePtr)
				assert.Equal(t, "pointer", *result.NamePtr)
				assert.NotNil(t, result.AgePtr)
				assert.Equal(t, 25, *result.AgePtr)
				assert.NotNil(t, result.ActivePtr)
				assert.True(t, *result.ActivePtr)
				assert.NotNil(t, result.ScorePtr)
				assert.Equal(t, 88.5, *result.ScorePtr)
			},
		},
		{
			name: "scan struct with nil pointer fields",
			rowData: map[string]any{
				"id":         1,
				"name_ptr":   nil,
				"age_ptr":    nil,
				"active_ptr": nil,
				"score_ptr":  nil,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Nil(t, result.NamePtr)
				assert.Nil(t, result.AgePtr)
				assert.Nil(t, result.ActivePtr)
				assert.Nil(t, result.ScorePtr)
			},
		},
		{
			name: "scan with partial data",
			rowData: map[string]any{
				"id":   1,
				"name": "partial",
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, "partial", result.Name)
				assert.Equal(t, 0, result.UserID)
				assert.Empty(t, result.UserName)
			},
		},
		{
			name:    "scan with empty data",
			rowData: map[string]any{},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 0, result.ID)
				assert.Empty(t, result.Name)
			},
		},
		{
			name: "scan with extra fields in data",
			rowData: map[string]any{
				"id":          1,
				"name":        "test",
				"extra_field": "ignored",
				"another":     42,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, "test", result.Name)
			},
		},
		{
			name: "scan with snake_case to CamelCase",
			rowData: map[string]any{
				"user_id":   10,
				"user_name": "john",
				"agent_id":  "agent123",
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 10, result.UserID)
				assert.Equal(t, "john", result.UserName)
				assert.Equal(t, "agent123", result.AgentId)
			},
		},
		{
			name: "scan with unexported field",
			rowData: map[string]any{
				"id":         1,
				"unexported": "ignored",
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Empty(t, result.unexported)
			},
		},
		{
			name: "scan empty slice",
			rowData: map[string]any{
				"id":    1,
				"slice": `[]`,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.Empty(t, result.Slice)
			},
		},
		{
			name: "scan empty map",
			rowData: map[string]any{
				"id":  1,
				"map": `{}`,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 1, result.ID)
				assert.NotNil(t, result.Map)
				assert.Empty(t, result.Map)
			},
		},
		{
			name: "scan with zero values",
			rowData: map[string]any{
				"id":     0,
				"name":   "",
				"active": false,
				"score":  0.0,
			},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*TestStruct)
				assert.Equal(t, 0, result.ID)
				assert.Empty(t, result.Name)
				assert.False(t, result.Active)
				assert.Equal(t, 0.0, result.Score)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, tt.rowErr)
			err := row.Scan(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, tt.target)
				}
			}
		})
	}
}

func TestScan_ToCarbon(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		assertion func(t *testing.T, target any)
	}{
		{
			name: "convert time.Time to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateTimeMilli",
			target: &struct {
				CreatedAt carbon.DateTimeMilli
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateTimeMicro",
			target: &struct {
				CreatedAt carbon.DateTimeMicro
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateTimeNano",
			target: &struct {
				CreatedAt carbon.DateTimeNano
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.Date",
			target: &struct {
				CreatedAt carbon.Date
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.Date
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateMilli",
			target: &struct {
				CreatedAt carbon.DateMilli
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateMicro",
			target: &struct {
				CreatedAt carbon.DateMicro
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.DateNano",
			target: &struct {
				CreatedAt carbon.DateNano
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.Timestamp",
			target: &struct {
				CreatedAt carbon.Timestamp
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.Timestamp
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.TimestampMilli",
			target: &struct {
				CreatedAt carbon.TimestampMilli
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.TimestampMicro",
			target: &struct {
				CreatedAt carbon.TimestampMicro
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert time.Time to carbon.TimestampNano",
			target: &struct {
				CreatedAt carbon.TimestampNano
			}{},
			rowData: map[string]any{"created_at": now},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateTimeMilli",
			target: &struct {
				CreatedAt carbon.DateTimeMilli
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateTimeMicro",
			target: &struct {
				CreatedAt carbon.DateTimeMicro
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateTimeNano",
			target: &struct {
				CreatedAt carbon.DateTimeNano
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTimeNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.Date",
			target: &struct {
				CreatedAt carbon.Date
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.Date
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateMilli",
			target: &struct {
				CreatedAt carbon.DateMilli
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateMicro",
			target: &struct {
				CreatedAt carbon.DateMicro
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.DateNano",
			target: &struct {
				CreatedAt carbon.DateNano
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.Timestamp",
			target: &struct {
				CreatedAt carbon.Timestamp
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.Timestamp
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.TimestampMilli",
			target: &struct {
				CreatedAt carbon.TimestampMilli
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampMilli
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.TimestampMicro",
			target: &struct {
				CreatedAt carbon.TimestampMicro
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampMicro
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
		{
			name: "convert string to carbon.TimestampNano",
			target: &struct {
				CreatedAt carbon.TimestampNano
			}{},
			rowData: map[string]any{"created_at": nowStr},
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.TimestampNano
				})
				assert.NotNil(t, result.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			err := row.Scan(tt.target)
			assert.NoError(t, err)
			if tt.assertion != nil {
				tt.assertion(t, tt.target)
			}
		})
	}
}

func TestScan_ToDeletedAt(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	tests := []struct {
		name      string
		rowData   map[string]any
		wantValid bool
		assertion func(t *testing.T, result gorm.DeletedAt)
	}{
		{
			name:      "convert time.Time to gorm.DeletedAt",
			rowData:   map[string]any{"deleted_at": now},
			wantValid: true,
			assertion: func(t *testing.T, result gorm.DeletedAt) {
				assert.True(t, result.Valid)
				assert.False(t, result.Time.IsZero())
			},
		},
		{
			name:      "convert string to gorm.DeletedAt",
			rowData:   map[string]any{"deleted_at": nowStr},
			wantValid: true,
			assertion: func(t *testing.T, result gorm.DeletedAt) {
				assert.True(t, result.Valid)
				assert.False(t, result.Time.IsZero())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				DeletedAt gorm.DeletedAt
			}
			row := NewRow(tt.rowData, nil)
			var result TestStruct
			err := row.Scan(&result)
			assert.NoError(t, err)
			if tt.assertion != nil {
				tt.assertion(t, result.DeletedAt)
			}
		})
	}
}

func TestScan_ToMap(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		wantErr   bool
		assertion func(t *testing.T, result any)
	}{
		// Empty string cases
		{
			name: "empty string to map[string]any",
			target: &struct {
				Metadata map[string]any
			}{},
			rowData: map[string]any{"metadata": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Metadata map[string]any
				})
				assert.NotNil(t, target.Metadata)
				assert.Empty(t, target.Metadata)
				assert.Equal(t, 0, len(target.Metadata))
			},
		},
		{
			name: "empty string to map[string]string",
			target: &struct {
				Labels map[string]string
			}{},
			rowData: map[string]any{"labels": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Labels map[string]string
				})
				assert.NotNil(t, target.Labels)
				assert.Empty(t, target.Labels)
				assert.Equal(t, 0, len(target.Labels))
			},
		},
		{
			name: "empty string to map[string]int",
			target: &struct {
				Counters map[string]int
			}{},
			rowData: map[string]any{"counters": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Counters map[string]int
				})
				assert.NotNil(t, target.Counters)
				assert.Empty(t, target.Counters)
				assert.Equal(t, 0, len(target.Counters))
			},
		},
		// Nil value cases
		{
			name: "nil to map[string]any",
			target: &struct {
				Metadata map[string]any
			}{},
			rowData: map[string]any{"metadata": nil},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Metadata map[string]any
				})
				assert.Nil(t, target.Metadata)
			},
		},
		{
			name: "nil to map[string]string",
			target: &struct {
				Labels map[string]string
			}{},
			rowData: map[string]any{"labels": nil},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Labels map[string]string
				})
				assert.Nil(t, target.Labels)
			},
		},
		// Valid JSON string cases
		{
			name: "json string to map[string]any",
			target: &struct {
				Metadata map[string]any
			}{},
			rowData: map[string]any{"metadata": `{"key": "value", "count": 10}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Metadata map[string]any
				})
				assert.Equal(t, "value", target.Metadata["key"])
				assert.Equal(t, float64(10), target.Metadata["count"])
			},
		},
		{
			name: "json string to map[string]string",
			target: &struct {
				Labels map[string]string
			}{},
			rowData: map[string]any{"labels": `{"env": "prod", "version": "1.0"}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Labels map[string]string
				})
				assert.Equal(t, "prod", target.Labels["env"])
				assert.Equal(t, "1.0", target.Labels["version"])
			},
		},
		{
			name: "json string to map[string]int",
			target: &struct {
				Counters map[string]int
			}{},
			rowData: map[string]any{"counters": `{"views": 100, "clicks": 50}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Counters map[string]int
				})
				assert.Equal(t, 100, target.Counters["views"])
				assert.Equal(t, 50, target.Counters["clicks"])
			},
		},
		{
			name: "json string to map[string]float64",
			target: &struct {
				Rates map[string]float64
			}{},
			rowData: map[string]any{"rates": `{"usd": 1.0, "eur": 0.85}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Rates map[string]float64
				})
				assert.Equal(t, 1.0, target.Rates["usd"])
				assert.Equal(t, 0.85, target.Rates["eur"])
			},
		},
		{
			name: "json string to map[string]bool",
			target: &struct {
				Flags map[string]bool
			}{},
			rowData: map[string]any{"flags": `{"enabled": true, "debug": false}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Flags map[string]bool
				})
				assert.Equal(t, true, target.Flags["enabled"])
				assert.Equal(t, false, target.Flags["debug"])
			},
		},
		{
			name: "json string to map with mixed types",
			target: &struct {
				Data map[string]any
			}{},
			rowData: map[string]any{"data": `{"name": "test", "age": 30, "active": true, "score": 98.5}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Data map[string]any
				})
				assert.Equal(t, "test", target.Data["name"])
				assert.Equal(t, float64(30), target.Data["age"])
				assert.Equal(t, true, target.Data["active"])
				assert.Equal(t, 98.5, target.Data["score"])
			},
		},
		{
			name: "json string to map with null values",
			target: &struct {
				Data map[string]any
			}{},
			rowData: map[string]any{"data": `{"key1": "value", "key2": null}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Data map[string]any
				})
				assert.Equal(t, "value", target.Data["key1"])
				assert.Nil(t, target.Data["key2"])
			},
		},

		// Single key-value cases
		{
			name: "single key-value map",
			target: &struct {
				Config map[string]string
			}{},
			rowData: map[string]any{"config": `{"theme": "dark"}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Config map[string]string
				})
				assert.Len(t, target.Config, 1)
				assert.Equal(t, "dark", target.Config["theme"])
			},
		},
		// Nested map cases
		{
			name: "nested maps",
			target: &struct {
				Config map[string]any
			}{},
			rowData: map[string]any{"config": `{"database": {"host": "localhost", "port": 5432}, "cache": {"ttl": 300}}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Config map[string]any
				})
				db := target.Config["database"].(map[string]any)
				assert.Equal(t, "localhost", db["host"])
				assert.Equal(t, float64(5432), db["port"])
				cache := target.Config["cache"].(map[string]any)
				assert.Equal(t, float64(300), cache["ttl"])
			},
		},
		// Map with array values
		{
			name: "map with array values",
			target: &struct {
				Data map[string]any
			}{},
			rowData: map[string]any{"data": `{"tags": ["go", "test"], "scores": [95, 88, 92]}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Data map[string]any
				})
				tags := target.Data["tags"].([]any)
				assert.Equal(t, "go", tags[0])
				assert.Equal(t, "test", tags[1])
				scores := target.Data["scores"].([]any)
				assert.Equal(t, float64(95), scores[0])
				assert.Equal(t, float64(88), scores[1])
				assert.Equal(t, float64(92), scores[2])
			},
		},
		// Large map case
		{
			name: "large map",
			target: &struct {
				Data map[string]int
			}{},
			rowData: map[string]any{"data": `{"k1":1,"k2":2,"k3":3,"k4":4,"k5":5,"k6":6,"k7":7,"k8":8,"k9":9,"k10":10}`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Data map[string]int
				})
				assert.Len(t, target.Data, 10)
				assert.Equal(t, 1, target.Data["k1"])
				assert.Equal(t, 10, target.Data["k10"])
			},
		},
		// Error cases
		{
			name: "invalid json string to map",
			target: &struct {
				Data map[string]any
			}{},
			rowData: map[string]any{"data": `{invalid json}`},
			wantErr: true,
		},
		{
			name: "non-object json to map",
			target: &struct {
				Data map[string]any
			}{},
			rowData: map[string]any{"data": `["array", "not", "object"]`},
			wantErr: true,
		},
		{
			name: "type mismatch in json object",
			target: &struct {
				Counters map[string]int
			}{},
			rowData: map[string]any{"counters": `{"key": "not_an_int"}`},
			wantErr: true,
		},
		// Multiple map fields
		{
			name: "multiple map fields",
			target: &struct {
				Labels   map[string]string
				Counters map[string]int
				Flags    map[string]bool
			}{},
			rowData: map[string]any{
				"labels":   `{"env": "prod"}`,
				"counters": `{"views": 100}`,
				"flags":    `{"enabled": true}`,
			},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Labels   map[string]string
					Counters map[string]int
					Flags    map[string]bool
				})
				assert.Equal(t, "prod", target.Labels["env"])
				assert.Equal(t, 100, target.Counters["views"])
				assert.Equal(t, true, target.Flags["enabled"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			err := row.Scan(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, tt.target)
				}
			}
		})
	}
}

func TestScan_ToScanner(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		wantErr   bool
		assertion func(t *testing.T, target any)
	}{
		{
			name: "same type should pass through - carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": carbon.Now()},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "same type should pass through - Json",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": Json{A: "test", B: 123}},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "test", result.Data.A)
				assert.Equal(t, 123, result.Data.B)
			},
		},
		{
			name: "scan valid json string to Json struct",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": `{"a": "test", "b": 42, "c": ["x", "y", "z"], "d": [1, 2, 3]}`},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "test", result.Data.A)
				assert.Equal(t, 42, result.Data.B)
				assert.Equal(t, []string{"x", "y", "z"}, result.Data.C)
				assert.Equal(t, []int{1, 2, 3}, result.Data.D)
			},
		},
		{
			name: "scan json with empty arrays",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": `{"a": "empty", "b": 0, "c": [], "d": []}`},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "empty", result.Data.A)
				assert.Equal(t, 0, result.Data.B)
				assert.Empty(t, result.Data.C)
				assert.Empty(t, result.Data.D)
			},
		},
		{
			name: "scan json bytes to Json struct",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": []byte(`{"a": "bytes", "b": 99, "c": ["a"], "d": [5]}`)},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "bytes", result.Data.A)
				assert.Equal(t, 99, result.Data.B)
				assert.Equal(t, []string{"a"}, result.Data.C)
				assert.Equal(t, []int{5}, result.Data.D)
			},
		},
		{
			name: "scan gorm.DeletedAt with nil value",
			target: &struct {
				DeletedAt gorm.DeletedAt
			}{},
			rowData: map[string]any{"deleted_at": nil},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					DeletedAt gorm.DeletedAt
				})
				assert.False(t, result.DeletedAt.Valid)
			},
		},
		{
			name: "scan multiple scanner fields",
			target: &struct {
				Data1 Json
				Data2 Json
			}{},
			rowData: map[string]any{
				"data1": `{"a": "first", "b": 1, "c": ["a"], "d": [1]}`,
				"data2": `{"a": "second", "b": 2, "c": ["b"], "d": [2]}`,
			},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data1 Json
					Data2 Json
				})
				assert.Equal(t, "first", result.Data1.A)
				assert.Equal(t, 1, result.Data1.B)
				assert.Equal(t, "second", result.Data2.A)
				assert.Equal(t, 2, result.Data2.B)
			},
		},
		{
			name: "convert []uint8 to custom scanner type - Json",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": []uint8(`{"a": "uint8", "b": 999}`)},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "uint8", result.Data.A)
				assert.Equal(t, 999, result.Data.B)
			},
		},
		{
			name: "empty string should return zero value",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": ""},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "", result.Data.A)
				assert.Equal(t, 0, result.Data.B)
			},
		},
		{
			name: "nil should return zero value",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": nil},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, "", result.Data.A)
				assert.Equal(t, 0, result.Data.B)
			},
		},
		{
			name: "string to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": now.Format(time.RFC3339)},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "[]byte to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": []byte(now.Format(time.RFC3339))},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "[]uint8 to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": []uint8(now.Format(time.RFC3339))},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "invalid json string should error",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": `{invalid json}`},
			wantErr: true,
		},
		{
			name: "time.Time should skip ToScannerHookFunc",
			target: &struct {
				CreatedAt time.Time
			}{},
			rowData: map[string]any{"created_at": now.Format(time.RFC3339)},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt time.Time
				})
				// Should be handled by ToTimeHookFunc, not ToScannerHookFunc
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "gorm.DeletedAt should skip ToScannerHookFunc",
			target: &struct {
				DeletedAt gorm.DeletedAt
			}{},
			rowData: map[string]any{"deleted_at": now.Format(time.RFC3339)},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					DeletedAt gorm.DeletedAt
				})
				// Should be handled by ToDeletedAtHookFunc, not ToScannerHookFunc
				assert.True(t, result.DeletedAt.Valid)
			},
		},
		{
			name: "time.Time as source should be processed by ToScannerHookFunc",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": now},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				// When source is time.Time, f != reflect.TypeOf(time.Time{}) is false,
				// so ToScannerHookFunc does NOT skip and continues processing
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "nil data to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": nil},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				// Carbon zero value is not nil and may not return IsZero() as true
				// Just verify no error occurred
				assert.NotNil(t, result)
			},
		},
		{
			name: "empty string to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": ""},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					CreatedAt carbon.DateTime
				})
				// Carbon zero value is not nil and may not return IsZero() as true
				// Just verify no error occurred
				assert.NotNil(t, result)
			},
		},
		{
			name: "nil data to Json",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": nil},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, Json{}, result.Data)
			},
		},
		{
			name: "empty string to Json",
			target: &struct {
				Data Json
			}{},
			rowData: map[string]any{"data": ""},
			wantErr: false,
			assertion: func(t *testing.T, target any) {
				result := target.(*struct {
					Data Json
				})
				assert.Equal(t, Json{}, result.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			err := row.Scan(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, tt.target)
				}
			}
		})
	}
}

func TestScan_ToSlice(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		wantErr   bool
		assertion func(t *testing.T, result any)
	}{
		// Empty string cases
		{
			name: "empty string to string slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags []string
				})
				assert.NotNil(t, target.Tags)
				assert.Empty(t, target.Tags)
				assert.Equal(t, 0, len(target.Tags))
			},
		},
		{
			name: "empty string to int slice",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Numbers []int
				})
				assert.NotNil(t, target.Numbers)
				assert.Empty(t, target.Numbers)
				assert.Equal(t, 0, len(target.Numbers))
			},
		},
		{
			name: "empty string to struct slice",
			target: &struct {
				Items []Json
			}{},
			rowData: map[string]any{"items": ""},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Items []Json
				})
				assert.NotNil(t, target.Items)
				assert.Empty(t, target.Items)
				assert.Equal(t, 0, len(target.Items))
			},
		},
		// Nil value cases
		{
			name: "nil to string slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": nil},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags []string
				})
				assert.Nil(t, target.Tags)
			},
		},
		{
			name: "nil to int slice",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": nil},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Numbers []int
				})
				assert.Nil(t, target.Numbers)
			},
		},
		// Valid JSON string cases
		{
			name: "json string to string slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": `["a", "b", "c"]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags []string
				})
				assert.Equal(t, []string{"a", "b", "c"}, target.Tags)
			},
		},
		{
			name: "json string to int slice",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": `[1, 2, 3, 4, 5]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Numbers []int
				})
				assert.Equal(t, []int{1, 2, 3, 4, 5}, target.Numbers)
			},
		},
		{
			name: "json string to float slice",
			target: &struct {
				Values []float64
			}{},
			rowData: map[string]any{"values": `[1.1, 2.2, 3.3]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Values []float64
				})
				assert.Equal(t, []float64{1.1, 2.2, 3.3}, target.Values)
			},
		},
		{
			name: "json string to bool slice",
			target: &struct {
				Flags []bool
			}{},
			rowData: map[string]any{"flags": `[true, false, true]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Flags []bool
				})
				assert.Equal(t, []bool{true, false, true}, target.Flags)
			},
		},
		{
			name: "json string to struct slice",
			target: &struct {
				Items []Json
			}{},
			rowData: map[string]any{"items": `[{"a": "first", "b": 1}, {"a": "second", "b": 2}]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Items []Json
				})
				assert.Len(t, target.Items, 2)
				assert.Equal(t, "first", target.Items[0].A)
				assert.Equal(t, 1, target.Items[0].B)
				assert.Equal(t, "second", target.Items[1].A)
				assert.Equal(t, 2, target.Items[1].B)
			},
		},
		{
			name: "json string to any slice",
			target: &struct {
				Data []any
			}{},
			rowData: map[string]any{"data": `["string", 123, true, null]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Data []any
				})
				assert.Len(t, target.Data, 4)
				assert.Equal(t, "string", target.Data[0])
				assert.Equal(t, float64(123), target.Data[1])
				assert.Equal(t, true, target.Data[2])
				assert.Nil(t, target.Data[3])
			},
		},

		// Single element cases
		{
			name: "single element string slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": `["single"]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags []string
				})
				assert.Equal(t, []string{"single"}, target.Tags)
			},
		},
		{
			name: "single element int slice",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": `[42]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Numbers []int
				})
				assert.Equal(t, []int{42}, target.Numbers)
			},
		},
		// Nested slice cases
		{
			name: "nested string slices",
			target: &struct {
				Matrix [][]string
			}{},
			rowData: map[string]any{"matrix": `[["a", "b"], ["c", "d"]]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Matrix [][]string
				})
				assert.Len(t, target.Matrix, 2)
				assert.Equal(t, []string{"a", "b"}, target.Matrix[0])
				assert.Equal(t, []string{"c", "d"}, target.Matrix[1])
			},
		},
		{
			name: "nested int slices",
			target: &struct {
				Matrix [][]int
			}{},
			rowData: map[string]any{"matrix": `[[1, 2], [3, 4]]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Matrix [][]int
				})
				assert.Len(t, target.Matrix, 2)
				assert.Equal(t, []int{1, 2}, target.Matrix[0])
				assert.Equal(t, []int{3, 4}, target.Matrix[1])
			},
		},
		// Large slice case
		{
			name: "large int slice",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": `[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]`},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Numbers []int
				})
				assert.Len(t, target.Numbers, 20)
				assert.Equal(t, 1, target.Numbers[0])
				assert.Equal(t, 20, target.Numbers[19])
			},
		},
		// Error cases
		{
			name: "invalid json string to slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": `[invalid json`},
			wantErr: true,
		},
		{
			name: "non-array json to slice",
			target: &struct {
				Tags []string
			}{},
			rowData: map[string]any{"tags": `{"key": "value"}`},
			wantErr: true,
		},
		{
			name: "type mismatch in json array",
			target: &struct {
				Numbers []int
			}{},
			rowData: map[string]any{"numbers": `["not", "numbers"]`},
			wantErr: true,
		},
		// Multiple slice fields
		{
			name: "multiple slice fields",
			target: &struct {
				Tags    []string
				Numbers []int
				Flags   []bool
			}{},
			rowData: map[string]any{
				"tags":    `["a", "b"]`,
				"numbers": `[1, 2]`,
				"flags":   `[true, false]`,
			},
			wantErr: false,
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags    []string
					Numbers []int
					Flags   []bool
				})
				assert.Equal(t, []string{"a", "b"}, target.Tags)
				assert.Equal(t, []int{1, 2}, target.Numbers)
				assert.Equal(t, []bool{true, false}, target.Flags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			err := row.Scan(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, tt.target)
				}
			}
		})
	}
}

func TestScan_ToString(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	tests := []struct {
		name      string
		rowData   map[string]any
		wantValue string
	}{
		{
			name:      "convert []uint8 to string",
			rowData:   map[string]any{"name": []uint8("test")},
			wantValue: "test",
		},
		{
			name:      "keep string as is",
			rowData:   map[string]any{"name": "test"},
			wantValue: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			var result TestStruct
			err := row.Scan(&result)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, result.Name)
		})
	}
}

func TestScan_ToTime(t *testing.T) {
	type TestStruct struct {
		CreatedAt time.Time
	}

	now := time.Now()
	nowRFC3339 := now.Format(time.RFC3339)

	tests := []struct {
		name      string
		rowData   map[string]any
		wantErr   bool
		assertion func(t *testing.T, result TestStruct)
	}{
		{
			name:    "convert string to time",
			rowData: map[string]any{"created_at": nowRFC3339},
			wantErr: false,
			assertion: func(t *testing.T, result TestStruct) {
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:    "convert float64 milliseconds to time",
			rowData: map[string]any{"created_at": float64(1609459200000)},
			wantErr: false,
			assertion: func(t *testing.T, result TestStruct) {
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:    "convert int64 milliseconds to time",
			rowData: map[string]any{"created_at": int64(1609459200000)},
			wantErr: false,
			assertion: func(t *testing.T, result TestStruct) {
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:    "invalid string format",
			rowData: map[string]any{"created_at": "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := NewRow(tt.rowData, nil)
			var result TestStruct
			err := row.Scan(&result)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, result)
				}
			}
		})
	}
}
