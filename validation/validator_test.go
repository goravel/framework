package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gookit/validate"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
)

func TestBind_Rule(t *testing.T) {
	type Data struct {
		A              string                  `form:"a" json:"a"`
		B              int                     `form:"b" json:"b"`
		File           *multipart.FileHeader   `form:"file" json:"file"`
		Files          []*multipart.FileHeader `form:"files" json:"files"`
		Ages           []int                   `form:"ages" json:"ages"`
		Names          []string                `form:"names" json:"names"`
		Carbon         *carbon.Carbon          `form:"carbon" json:"carbon"`
		DateTime       *carbon.DateTime        `form:"date_time" json:"date_time"`
		DateTimeMilli  *carbon.DateTimeMilli   `form:"date_time_milli" json:"date_time_milli"`
		DateTimeMicro  *carbon.DateTimeMicro   `form:"date_time_micro" json:"date_time_micro"`
		DateTimeNano   *carbon.DateTimeNano    `form:"date_time_nano" json:"date_time_nano"`
		Date           *carbon.Date            `form:"date" json:"date"`
		DateMilli      *carbon.DateMilli       `form:"date_milli" json:"date_milli"`
		DateMicro      *carbon.DateMicro       `form:"date_micro" json:"date_micro"`
		DateNano       *carbon.DateNano        `form:"date_nano" json:"date_nano"`
		Timestamp      *carbon.Timestamp       `form:"timestamp" json:"timestamp"`
		TimestampMilli *carbon.TimestampMilli  `form:"timestamp_milli" json:"timestamp_milli"`
		TimestampMicro *carbon.TimestampMicro  `form:"timestamp_micro" json:"timestamp_micro"`
		TimestampNano  *carbon.TimestampNano   `form:"timestamp_nano" json:"timestamp_nano"`
		Time           *time.Time              `form:"time" json:"time"`
	}

	tests := []struct {
		name   string
		data   validate.DataFace
		rules  map[string]string
		assert func(data Data)
	}{
		{
			name:  "data is map and key is lowercase",
			data:  validate.FromMap(map[string]any{"a": "aa", "b": "1"}),
			rules: map[string]string{"a": "required"},
			assert: func(data Data) {
				assert.Equal(t, "aa", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name:  "data is map and cast key",
			data:  validate.FromMap(map[string]any{"b": "1"}),
			rules: map[string]string{"b": "required"},
			assert: func(data Data) {
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name:  "data is map and key is uppercase",
			data:  validate.FromMap(map[string]any{"A": "aa"}),
			rules: map[string]string{"A": "required"},
			assert: func(data Data) {
				assert.Equal(t, "aa", data.A)
			},
		},
		{
			name: "data is struct",
			data: func() validate.DataFace {
				data, err := validate.FromStruct(&struct {
					A string
					B int
				}{
					A: "a",
					B: 1,
				})
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"A": "required"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name: "data is get request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodGet, "/?a=aa&&b=1", nil)
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required"},
			assert: func(data Data) {
				assert.Equal(t, "aa", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name: "data is post request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"a":"Goravel", "b": 1, "ages": [1, 2], "names": ["a", "b"]}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				age, exist := data.Get("ages")
				assert.True(t, exist)

				_, err = data.Set("ages", cast.ToIntSlice(age))
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required", "ages.*": "int", "names.*": "string"},
			assert: func(data Data) {
				assert.Equal(t, "Goravel", data.A)
				assert.Equal(t, 1, data.B)
				assert.Equal(t, []int{1, 2}, data.Ages)
				assert.Equal(t, []string{"a", "b"}, data.Names)
			},
		},
		{
			name: "data is post request with Carbon",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"carbon": "2024-07-04 10:00:52"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"carbon": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.Carbon.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": "2024-07-04 10:00:52"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime(string)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": "1720087252"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": 1720087252}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime(milli)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": 1720087252000}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime(micro)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": 1720087252000000}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTime(nano)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time": 1720087252000000000}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.DateTime.ToDateTimeString())
			},
		},
		{
			name: "data is post request with DateTimeMilli",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time_milli": "2024-07-04 10:00:52.123"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time_milli": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123", data.DateTimeMilli.ToDateTimeMilliString())
			},
		},
		{
			name: "data is post request with DateTimeMilli(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time_milli": 1720087252123}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time_milli": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123", data.DateTimeMilli.ToDateTimeMilliString())
			},
		},
		{
			name: "data is post request with DateTimeMicro",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time_micro": "2024-07-04 10:00:52.123456"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time_micro": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123456", data.DateTimeMicro.ToDateTimeMicroString())
			},
		},
		{
			name: "data is post request with DateTimeNano",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time_nano": "2024-07-04 10:00:52.123456789"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time_nano": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123456789", data.DateTimeNano.ToDateTimeNanoString())
			},
		},
		{
			name: "data is post request with DateTimeNano(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_time_nano": "1720087252123456789"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_time_nano": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123456789", data.DateTimeNano.ToDateTimeNanoString())
			},
		},
		{
			name: "data is post request with Date",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date": "2024-07-04"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04", data.Date.ToDateString())
			},
		},
		{
			name: "data is post request with Date(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date": 1720087252}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04", data.Date.ToDateString())
			},
		},
		{
			name: "data is post request with DateMilli",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_milli": "2024-07-04.123"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_milli": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123", data.DateMilli.ToDateMilliString())
			},
		},
		{
			name: "data is post request with DateMilli(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_milli": 1720087252123}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_milli": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123", data.DateMilli.ToDateMilliString())
			},
		},
		{
			name: "data is post request with DateMicro",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_micro": "2024-07-04.123456"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_micro": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123456", data.DateMicro.ToDateMicroString())
			},
		},
		{
			name: "data is post request with DateMicro(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_micro": 1720087252123456}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_micro": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123456", data.DateMicro.ToDateMicroString())
			},
		},
		{
			name: "data is post request with DateNano",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_nano": "2024-07-04.123456789"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_nano": "string"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123456789", data.DateNano.ToDateNanoString())
			},
		},
		{
			name: "data is post request with DateNano(int)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"date_nano": "1720087252123456789"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"date_nano": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04.123456789", data.DateNano.ToDateNanoString())
			},
		},
		{
			name: "data is post request with Timestamp",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"timestamp": 1720087252}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"timestamp": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52", data.Timestamp.ToDateTimeString())
			},
		},
		{
			name: "data is post request with TimestampMilli",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"timestamp_milli": 1720087252123}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"timestamp_milli": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123", data.TimestampMilli.ToDateTimeMilliString())
			},
		},
		{
			name: "data is post request with TimestampMicro",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"timestamp_micro": 1720087252123456}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"timestamp_micro": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123456", data.TimestampMicro.ToDateTimeMicroString())
			},
		},
		{
			name: "data is post request with TimestampNano",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"timestamp_nano": "1720087252123456789"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"timestamp_nano": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2024-07-04 10:00:52.123456789", data.TimestampNano.ToDateTimeNanoString())
			},
		},
		{
			name: "data is post request with Time",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"time": "2025-05-23 22:16:39"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2025-05-23 22:16:39", data.Time.Format("2006-01-02 15:04:05"))
			},
		},
		{
			name: "data is post request with Time(date)",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"time": "2025-05-23"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)

				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"time": "required"},
			assert: func(data Data) {
				assert.Equal(t, "2025-05-23", data.Time.Format("2006-01-02"))
			},
		},
		{
			name: "data is post request with body",
			data: func() validate.DataFace {
				request := buildRequest(t)
				data, err := validate.FromRequest(request, 1)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required", "file": "file"},
			assert: func(data Data) {
				request := buildRequest(t)
				_, file, err := request.FormFile("file")
				assert.Nil(t, err)

				assert.Equal(t, "aa", data.A)
				assert.NotNil(t, data.File)
				assert.Equal(t, file.Filename, data.File.Filename)
			},
		},
		{
			name: "data is post request with multiple files",
			data: func() validate.DataFace {
				request := buildRequestWithMultipleFiles(t)
				data, err := validate.FromRequest(request, 1)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required", "files": "file"},
			assert: func(data Data) {
				request := buildRequestWithMultipleFiles(t)
				_, file, err := request.FormFile("files")
				assert.Nil(t, err)

				assert.Equal(t, "aa", data.A)
				assert.Len(t, data.Files, 2)
				assert.Equal(t, file.Filename, data.Files[0].Filename)
				assert.Equal(t, file.Filename, data.Files[1].Filename)
			},
		},
	}

	validation := NewValidation()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.Make(test.data, test.rules)
			require.Nil(t, err)
			require.Nil(t, validator.Errors())

			var data Data
			err = validator.Bind(&data)
			require.Nil(t, err)

			test.assert(data)
		})
	}
}

func TestBind_Filter(t *testing.T) {
	type Data struct {
		A string `form:"a" json:"a"`
		B int    `form:"b" json:"b"`
	}

	tests := []struct {
		name    string
		data    validate.DataFace
		rules   map[string]string
		filters map[string]string
		assert  func(data Data)
	}{
		{
			name:    "data is map and key is lowercase",
			data:    validate.FromMap(map[string]any{"a": " a ", "b": "1"}),
			rules:   map[string]string{"a": "required", "b": "required"},
			filters: map[string]string{"a": "trim", "b": "int"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name:    "data is map and key is lowercase, a no rule but has filter, a should keep the original value.",
			data:    validate.FromMap(map[string]any{"a": "a", "b": " 1"}),
			rules:   map[string]string{"b": "required"},
			filters: map[string]string{"a": "upper", "b": "trim|int"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name: "data is struct",
			data: func() validate.DataFace {
				data, err := validate.FromStruct(&struct {
					A string
				}{
					A: " a ",
				})
				assert.Nil(t, err)

				return data
			}(),
			rules:   map[string]string{"A": "required"},
			filters: map[string]string{"A": "trim"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
			},
		},
		{
			name: "data is get request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodGet, "/?a= a &&b=1", nil)
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules:   map[string]string{"a": "required", "b": "required"},
			filters: map[string]string{"a": "trim"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
				assert.Equal(t, 1, data.B)
			},
		},
		{
			name: "data is post request with body",
			data: func() validate.DataFace {
				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)

				err := writer.WriteField("a", " a ")
				assert.Nil(t, err)
				assert.Nil(t, writer.Close())

				request, err := http.NewRequest(http.MethodPost, "/", payload)
				assert.Nil(t, err)
				request.Header.Set("Content-Type", writer.FormDataContentType())

				data, err := validate.FromRequest(request, 1)
				assert.Nil(t, err)

				return data
			}(),
			rules:   map[string]string{"a": "required", "file": "file"},
			filters: map[string]string{"a": "trim"},
			assert: func(data Data) {
				assert.Equal(t, "a", data.A)
			},
		},
	}

	validation := NewValidation()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.Make(test.data, test.rules, Filters(test.filters))
			require.Nil(t, err)
			require.Nil(t, validator.Errors())

			var data Data
			err = validator.Bind(&data)
			require.Nil(t, err)

			test.assert(data)
		})
	}
}

func TestFails(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		filters   map[string]string
		expectRes bool
	}{
		{
			describe: "false",
			data:     map[string]any{"a": "aa"},
			rules:    map[string]string{"a": "required"},
			filters:  map[string]string{},
		},
		{
			describe:  "true",
			data:      map[string]any{"b": "bb"},
			rules:     map[string]string{"a": "required"},
			filters:   map[string]string{},
			expectRes: true,
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			test.data,
			test.rules,
			Filters(test.filters),
		)
		assert.Nil(t, err)
		assert.Equal(t, test.expectRes, validator.Fails(), test.describe)
	}
}

func TestCastValue(t *testing.T) {
	date := "2024-07-04"
	dateTime := "2024-07-04 10:00:52"
	timestamp := int64(1720087252)
	timestampMilli := int64(1720087252000)
	timestampMicro := int64(1720087252000000)
	timestampNano := int64(1720087252000000000)

	type Data struct {
		String                string                 `form:"String" json:"String"`
		Int                   int                    `form:"Int" json:"Int"`
		Int8                  int8                   `form:"Int8" json:"Int8"`
		Int16                 int16                  `form:"Int16" json:"Int16"`
		Int32                 int32                  `form:"Int32" json:"Int32"`
		Int64                 int64                  `form:"Int64" json:"Int64"`
		Uint                  uint                   `form:"Uint" json:"Uint"`
		Uint8                 uint8                  `form:"Uint8" json:"Uint8"`
		Uint16                uint16                 `form:"Uint16" json:"Uint16"`
		Uint32                uint32                 `form:"Uint32" json:"Uint32"`
		Uint64                uint64                 `form:"Uint64" json:"Uint64"`
		Bool                  bool                   `form:"Bool" json:"Bool"`
		Float32               float32                `form:"Float32" json:"Float32"`
		Float64               float64                `form:"Float64" json:"Float64"`
		StringSlice           []string               `form:"StringSlice" json:"StringSlice"`
		IntSlice              []int                  `form:"IntSlice" json:"IntSlice"`
		BoolSlice             []bool                 `form:"BoolSlice" json:"BoolSlice"`
		FloatSlice            []float64              `form:"FloatSlice" json:"FloatSlice"`
		Map                   map[string]string      `form:"Map" json:"Map"`
		PointerCarbon         *carbon.Carbon         `form:"PointerCarbon" json:"PointerCarbon"`
		PointerDateTime       *carbon.DateTime       `form:"PointerDateTime" json:"PointerDateTime"`
		PointerDateTimeMilli  *carbon.DateTimeMilli  `form:"PointerDateTimeMilli" json:"PointerDateTimeMilli"`
		PointerDateTimeMicro  *carbon.DateTimeMicro  `form:"PointerDateTimeMicro" json:"PointerDateTimeMicro"`
		PointerDateTimeNano   *carbon.DateTimeNano   `form:"PointerDateTimeNano" json:"PointerDateTimeNano"`
		PointerDate           *carbon.Date           `form:"PointerDate" json:"PointerDate"`
		PointerDateMilli      *carbon.DateMilli      `form:"PointerDateMilli" json:"PointerDateMilli"`
		PointerDateMicro      *carbon.DateMicro      `form:"PointerDateMicro" json:"PointerDateMicro"`
		PointerDateNano       *carbon.DateNano       `form:"PointerDateNano" json:"PointerDateNano"`
		PointerTimestamp      *carbon.Timestamp      `form:"PointerTimestamp" json:"PointerTimestamp"`
		PointerTimestampMilli *carbon.TimestampMilli `form:"PointerTimestampMilli" json:"PointerTimestampMilli"`
		PointerTimestampMicro *carbon.TimestampMicro `form:"PointerTimestampMicro" json:"PointerTimestampMicro"`
		PointerTimestampNano  *carbon.TimestampNano  `form:"PointerTimestampNano" json:"PointerTimestampNano"`
		PointerTime           *time.Time             `form:"PointerTime" json:"PointerTime"`
		Carbon                carbon.Carbon          `form:"Carbon" json:"Carbon"`
		DateTime              carbon.DateTime        `form:"DateTime" json:"DateTime"`
		DateTimeMilli         carbon.DateTimeMilli   `form:"DateTimeMilli" json:"DateTimeMilli"`
		DateTimeMicro         carbon.DateTimeMicro   `form:"DateTimeMicro" json:"DateTimeMicro"`
		DateTimeNano          carbon.DateTimeNano    `form:"DateTimeNano" json:"DateTimeNano"`
		Date                  carbon.Date            `form:"Date" json:"Date"`
		DateMilli             carbon.DateMilli       `form:"DateMilli" json:"DateMilli"`
		DateMicro             carbon.DateMicro       `form:"DateMicro" json:"DateMicro"`
		DateNano              carbon.DateNano        `form:"DateNano" json:"DateNano"`
		Timestamp             carbon.Timestamp       `form:"Timestamp" json:"Timestamp"`
		TimestampMilli        carbon.TimestampMilli  `form:"TimestampMilli" json:"TimestampMilli"`
		TimestampMicro        carbon.TimestampMicro  `form:"TimestampMicro" json:"TimestampMicro"`
		TimestampNano         carbon.TimestampNano   `form:"TimestampNano" json:"TimestampNano"`
		Time                  time.Time              `form:"Time" json:"Time"`
	}

	wantData := Data{
		String:                "1",
		Int:                   1,
		Int8:                  2,
		Int16:                 3,
		Int32:                 4,
		Int64:                 5,
		Uint:                  6,
		Uint8:                 7,
		Uint16:                8,
		Uint32:                9,
		Uint64:                10,
		Bool:                  true,
		Float32:               11.11,
		Float64:               12.12,
		StringSlice:           []string{"1"},
		IntSlice:              []int{1},
		BoolSlice:             []bool{true, false},
		FloatSlice:            []float64{11.11, 12.12},
		Map:                   map[string]string{"a": "aa"},
		PointerCarbon:         carbon.Parse(dateTime),
		PointerDateTime:       carbon.NewDateTime(carbon.Parse(dateTime)),
		PointerDateTimeMilli:  carbon.NewDateTimeMilli(carbon.Parse(dateTime)),
		PointerDateTimeMicro:  carbon.NewDateTimeMicro(carbon.Parse(dateTime)),
		PointerDateTimeNano:   carbon.NewDateTimeNano(carbon.Parse(dateTime)),
		PointerDate:           carbon.NewDate(carbon.Parse(date)),
		PointerDateMilli:      carbon.NewDateMilli(carbon.Parse(date)),
		PointerDateMicro:      carbon.NewDateMicro(carbon.Parse(date)),
		PointerDateNano:       carbon.NewDateNano(carbon.Parse(date)),
		PointerTimestamp:      carbon.NewTimestamp(carbon.FromTimestamp(timestamp)),
		PointerTimestampMilli: carbon.NewTimestampMilli(carbon.FromTimestampMilli(timestampMilli)),
		PointerTimestampMicro: carbon.NewTimestampMicro(carbon.FromTimestampMicro(timestampMicro)),
		PointerTimestampNano:  carbon.NewTimestampNano(carbon.FromTimestampNano(timestampNano)),
		PointerTime:           convert.Pointer(carbon.NewDateTime(carbon.Parse(dateTime)).StdTime()),
		Carbon:                *carbon.Parse(dateTime),
		DateTime:              *carbon.NewDateTime(carbon.Parse(dateTime)),
		DateTimeMilli:         *carbon.NewDateTimeMilli(carbon.Parse(dateTime)),
		DateTimeMicro:         *carbon.NewDateTimeMicro(carbon.Parse(dateTime)),
		DateTimeNano:          *carbon.NewDateTimeNano(carbon.Parse(dateTime)),
		Date:                  *carbon.NewDate(carbon.Parse(date)),
		DateMilli:             *carbon.NewDateMilli(carbon.Parse(date)),
		DateMicro:             *carbon.NewDateMicro(carbon.Parse(date)),
		DateNano:              *carbon.NewDateNano(carbon.Parse(date)),
		Timestamp:             *carbon.NewTimestamp(carbon.FromTimestamp(timestamp)),
		TimestampMilli:        *carbon.NewTimestampMilli(carbon.FromTimestampMilli(timestampMilli)),
		TimestampMicro:        *carbon.NewTimestampMicro(carbon.FromTimestampMicro(timestampMicro)),
		TimestampNano:         *carbon.NewTimestampNano(carbon.FromTimestampNano(timestampNano)),
		Time:                  carbon.NewDateTime(carbon.Parse(dateTime)).StdTime(),
	}

	tests := []struct {
		name    string
		data    validate.DataFace
		wantErr error
	}{
		{
			name: "success with struct",
			data: func() validate.DataFace {
				body := &Data{
					String:                "1",
					Int:                   1,
					Int8:                  2,
					Int16:                 3,
					Int32:                 4,
					Int64:                 5,
					Uint:                  6,
					Uint8:                 7,
					Uint16:                8,
					Uint32:                9,
					Uint64:                10,
					Bool:                  true,
					Float32:               11.11,
					Float64:               12.12,
					StringSlice:           []string{"1"},
					IntSlice:              []int{1},
					BoolSlice:             []bool{true, false},
					FloatSlice:            []float64{11.11, 12.12},
					Map:                   map[string]string{"a": "aa"},
					PointerCarbon:         carbon.Parse(dateTime),
					PointerDateTime:       carbon.NewDateTime(carbon.Parse(dateTime)),
					PointerDateTimeMilli:  carbon.NewDateTimeMilli(carbon.Parse(dateTime)),
					PointerDateTimeMicro:  carbon.NewDateTimeMicro(carbon.Parse(dateTime)),
					PointerDateTimeNano:   carbon.NewDateTimeNano(carbon.Parse(dateTime)),
					PointerDate:           carbon.NewDate(carbon.Parse(dateTime)),
					PointerDateMilli:      carbon.NewDateMilli(carbon.Parse(dateTime)),
					PointerDateMicro:      carbon.NewDateMicro(carbon.Parse(dateTime)),
					PointerDateNano:       carbon.NewDateNano(carbon.Parse(dateTime)),
					PointerTimestamp:      carbon.NewTimestamp(carbon.Parse(dateTime)),
					PointerTimestampMilli: carbon.NewTimestampMilli(carbon.Parse(dateTime)),
					PointerTimestampMicro: carbon.NewTimestampMicro(carbon.Parse(dateTime)),
					PointerTimestampNano:  carbon.NewTimestampNano(carbon.Parse(dateTime)),
					PointerTime:           convert.Pointer(carbon.NewDateTime(carbon.Parse(dateTime)).StdTime()),
					Carbon:                *carbon.Parse(dateTime),
					DateTime:              *carbon.NewDateTime(carbon.Parse(dateTime)),
					DateTimeMilli:         *carbon.NewDateTimeMilli(carbon.Parse(dateTime)),
					DateTimeMicro:         *carbon.NewDateTimeMicro(carbon.Parse(dateTime)),
					DateTimeNano:          *carbon.NewDateTimeNano(carbon.Parse(dateTime)),
					Date:                  *carbon.NewDate(carbon.Parse(dateTime)),
					DateMilli:             *carbon.NewDateMilli(carbon.Parse(dateTime)),
					DateMicro:             *carbon.NewDateMicro(carbon.Parse(dateTime)),
					DateNano:              *carbon.NewDateNano(carbon.Parse(dateTime)),
					Timestamp:             *carbon.NewTimestamp(carbon.Parse(dateTime)),
					TimestampMilli:        *carbon.NewTimestampMilli(carbon.Parse(dateTime)),
					TimestampMicro:        *carbon.NewTimestampMicro(carbon.Parse(dateTime)),
					TimestampNano:         *carbon.NewTimestampNano(carbon.Parse(dateTime)),
					Time:                  carbon.NewDateTime(carbon.Parse(dateTime)).StdTime(),
				}
				jsonBytes, err := json.New().Marshal(body)
				assert.Nil(t, err)
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBytes))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
		},
		{
			name: "success with map",
			data: func() validate.DataFace {
				body := map[string]any{
					"String":                "1",
					"Int":                   "1",
					"Int8":                  "2",
					"Int16":                 "3",
					"Int32":                 "4",
					"Int64":                 "5",
					"Uint":                  "6",
					"Uint8":                 "7",
					"Uint16":                "8",
					"Uint32":                "9",
					"Uint64":                "10",
					"Bool":                  "true",
					"Float32":               "11.11",
					"Float64":               "12.12",
					"StringSlice":           []string{"1"},
					"IntSlice":              []string{"1"},
					"BoolSlice":             []string{"true", "false"},
					"FloatSlice":            []string{"11.11", "12.12"},
					"Map":                   map[string]string{"a": "aa"},
					"PointerCarbon":         dateTime,
					"PointerDateTime":       dateTime,
					"PointerDateTimeMilli":  dateTime,
					"PointerDateTimeMicro":  dateTime,
					"PointerDateTimeNano":   dateTime,
					"PointerDate":           date,
					"PointerDateMilli":      date,
					"PointerDateMicro":      date,
					"PointerDateNano":       date,
					"PointerTimestamp":      timestamp,
					"PointerTimestampMilli": timestampMilli,
					"PointerTimestampMicro": timestampMicro,
					"PointerTimestampNano":  timestampNano,
					"PointerTime":           dateTime,
					"Carbon":                dateTime,
					"DateTime":              dateTime,
					"DateTimeMilli":         dateTime,
					"DateTimeMicro":         dateTime,
					"DateTimeNano":          dateTime,
					"Date":                  date,
					"DateMilli":             date,
					"DateMicro":             date,
					"DateNano":              date,
					"Timestamp":             timestamp,
					"TimestampMilli":        timestampMilli,
					"TimestampMicro":        timestampMicro,
					"TimestampNano":         timestampNano,
					"Time":                  dateTime,
				}
				jsonBytes, err := json.New().Marshal(body)
				assert.Nil(t, err)
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBytes))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
		},
	}

	validation := NewValidation()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.Make(test.data, map[string]string{
				"String": "required",
			})
			assert.Nil(t, err)

			assert.False(t, validator.Fails())

			var data Data
			err = validator.Bind(&data)
			assert.Nil(t, test.wantErr, err)
			assert.Equal(t, wantData, data)
		})
	}
}

func TestCastCarbon(t *testing.T) {
	tests := []struct {
		name      string
		from      reflect.Value
		transform func(carbon *carbon.Carbon) any
		assert    func(result any)
	}{
		{
			name: "Happy path - length 10 string",
			from: reflect.ValueOf("2024-07-04"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewDate(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.Date{}, result)
				assert.Equal(t, "2024-07-04", result.(*carbon.Date).ToDateString())
			},
		},
		{
			name: "Happy path - length 10 int",
			from: reflect.ValueOf(1720087252),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestamp(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.Timestamp{}, result)
				assert.Equal(t, "2024-07-04 10:00:52", result.(*carbon.Timestamp).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 13 int",
			from: reflect.ValueOf(1720087252123),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(*carbon.TimestampMilli).ToDateTimeMilliString())
			},
		},
		{
			name: "Sad path - length 13 string",
			from: reflect.ValueOf("1720087252123"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(*carbon.TimestampMilli).ToDateTimeMilliString())
			},
		},
		{
			name: "Happy path - length 13 Y-m-d H",
			from: reflect.ValueOf("2024-07-04 10"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.Equal(t, "2024-07-04 10:00:00", result.(*carbon.TimestampMilli).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 16 int",
			from: reflect.ValueOf(1720087252123456),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampMicro(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampMicro{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456", result.(*carbon.TimestampMicro).ToDateTimeMicroString())
			},
		},
		{
			name: "Happy path - length 16 string",
			from: reflect.ValueOf("1720087252123456"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampMicro(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampMicro{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456", result.(*carbon.TimestampMicro).ToDateTimeMicroString())
			},
		},
		{
			name: "Happy path - length 16 Y-m-d H:i",
			from: reflect.ValueOf("2024-07-04 10:00"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewDateTime(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.DateTime{}, result)
				assert.Equal(t, "2024-07-04 10:00:00", result.(*carbon.DateTime).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 19 int",
			from: reflect.ValueOf(1720087252123456789),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampNano(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampNano{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456789", result.(*carbon.TimestampNano).ToDateTimeNanoString())
			},
		},
		{
			name: "Happy path - length 19 int",
			from: reflect.ValueOf("1720087252123456789"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewTimestampNano(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.TimestampNano{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456789", result.(*carbon.TimestampNano).ToDateTimeNanoString())
			},
		},
		{
			name: "Happy path - length 19 Y-m-d H:i:s",
			from: reflect.ValueOf("2024-07-04 10:00:52"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewDateTime(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.DateTime{}, result)
				assert.Equal(t, "2024-07-04 10:00:52", result.(*carbon.DateTime).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length other",
			from: reflect.ValueOf("2024-07-04 10:00:52.123"),
			transform: func(c *carbon.Carbon) any {
				return carbon.NewDateTimeMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, &carbon.DateTimeMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(*carbon.DateTimeMilli).ToDateTimeMilliString())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.assert(castCarbon(test.from, test.transform))
		})
	}
}

func buildRequest(t *testing.T) *http.Request {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	err := writer.WriteField("a", "aa")
	assert.Nil(t, err)

	logo, err := os.Open("../logo.png")
	assert.Nil(t, err)

	defer logo.Close()
	part1, err := writer.CreateFormFile("file", filepath.Base("../logo.png"))
	assert.Nil(t, err)

	_, err = io.Copy(part1, logo)
	assert.Nil(t, err)
	assert.Nil(t, writer.Close())

	request, err := http.NewRequest(http.MethodPost, "/", payload)
	assert.Nil(t, err)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request
}

func buildRequestWithMultipleFiles(t *testing.T) *http.Request {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	err := writer.WriteField("a", "aa")
	assert.Nil(t, err)

	logo1, err := os.Open("../logo.png")
	assert.Nil(t, err)

	defer logo1.Close()
	part1, err := writer.CreateFormFile("files", filepath.Base("../logo.png"))
	assert.Nil(t, err)

	_, err = io.Copy(part1, logo1)
	assert.Nil(t, err)

	logo2, err := os.Open("../logo.png")
	assert.Nil(t, err)

	defer logo2.Close()
	part2, err := writer.CreateFormFile("files", filepath.Base("../logo.png"))
	assert.Nil(t, err)

	_, err = io.Copy(part2, logo2)
	assert.Nil(t, err)

	assert.Nil(t, writer.Close())

	request, err := http.NewRequest(http.MethodPost, "/", payload)
	assert.Nil(t, err)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request
}
