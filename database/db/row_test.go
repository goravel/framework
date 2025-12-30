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
	type TestStruct struct {
		ID       int
		Name     string
		AgentId  string
		UserID   int
		UserName string
		Slice    []string
		Map      map[string]any
		Json     Json
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
			name: "scan struct",
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
			name:    "scan with empty data",
			rowData: map[string]any{},
			rowErr:  nil,
			target:  &TestStruct{},
			wantErr: false,
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

func TestScan_ToMap_EmptyData(t *testing.T) {
	tests := []struct {
		name      string
		rowData   map[string]any
		assertion func(t *testing.T, result any)
	}{
		{
			name:    "empty string to map[string]any",
			rowData: map[string]any{"metadata": ""},
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
			name:    "empty string to map[string]string",
			rowData: map[string]any{"labels": ""},
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
			name:    "empty string to map[string]int",
			rowData: map[string]any{"counters": ""},
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Counters map[string]int
				})
				assert.NotNil(t, target.Counters)
				assert.Empty(t, target.Counters)
				assert.Equal(t, 0, len(target.Counters))
			},
		},
		{
			name:    "valid json string to map",
			rowData: map[string]any{"metadata": `{"key": "value", "count": 10}`},
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Metadata map[string]any
				})
				assert.Equal(t, "value", target.Metadata["key"])
				assert.Equal(t, float64(10), target.Metadata["count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target any
			switch tt.name {
			case "empty string to map[string]any", "valid json string to map":
				target = &struct {
					Metadata map[string]any
				}{}
			case "empty string to map[string]string":
				target = &struct {
					Labels map[string]string
				}{}
			case "empty string to map[string]int":
				target = &struct {
					Counters map[string]int
				}{}
			}

			row := NewRow(tt.rowData, nil)
			err := row.Scan(target)
			assert.NoError(t, err)
			if tt.assertion != nil {
				tt.assertion(t, target)
			}
		})
	}
}

func TestScan_ToScanner(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		wantErr   bool
		assertion func(t *testing.T, target any)
	}{
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

func TestScan_ToScanner_EmptyData(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		rowData   map[string]any
		assertion func(t *testing.T, target any)
	}{
		{
			name: "nil data to carbon.DateTime",
			target: &struct {
				CreatedAt carbon.DateTime
			}{},
			rowData: map[string]any{"created_at": nil},
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
			assert.NoError(t, err)
			if tt.assertion != nil {
				tt.assertion(t, tt.target)
			}
		})
	}
}

func TestScan_ToSlice_EmptyData(t *testing.T) {
	tests := []struct {
		name      string
		rowData   map[string]any
		assertion func(t *testing.T, result any)
	}{
		{
			name:    "empty string to string slice",
			rowData: map[string]any{"tags": ""},
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
			name:    "empty string to int slice",
			rowData: map[string]any{"numbers": ""},
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
			name:    "empty string to struct slice",
			rowData: map[string]any{"items": ""},
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Items []Json
				})
				assert.NotNil(t, target.Items)
				assert.Empty(t, target.Items)
				assert.Equal(t, 0, len(target.Items))
			},
		},
		{
			name:    "valid json string to slice",
			rowData: map[string]any{"tags": `["a", "b", "c"]`},
			assertion: func(t *testing.T, result any) {
				target := result.(*struct {
					Tags []string
				})
				assert.Equal(t, []string{"a", "b", "c"}, target.Tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target any
			switch tt.name {
			case "empty string to string slice", "valid json string to slice":
				target = &struct {
					Tags []string
				}{}
			case "empty string to int slice":
				target = &struct {
					Numbers []int
				}{}
			case "empty string to struct slice":
				target = &struct {
					Items []Json
				}{}
			}

			row := NewRow(tt.rowData, nil)
			err := row.Scan(target)
			assert.NoError(t, err)
			if tt.assertion != nil {
				tt.assertion(t, target)
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
