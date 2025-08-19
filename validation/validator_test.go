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

	"github.com/gookit/validate"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/carbon"
)

func TestBind_Rule(t *testing.T) {
	type Data struct {
		A              string                 `form:"a" json:"a"`
		B              int                    `form:"b" json:"b"`
		File           *multipart.FileHeader  `form:"file" json:"file"`
		Ages           []int                  `form:"ages" json:"ages"`
		Names          []string               `form:"names" json:"names"`
		Carbon         *carbon.Carbon         `form:"carbon" json:"carbon"`
		DateTime       *carbon.DateTime       `form:"date_time" json:"date_time"`
		DateTimeMilli  *carbon.DateTimeMilli  `form:"date_time_milli" json:"date_time_milli"`
		DateTimeMicro  *carbon.DateTimeMicro  `form:"date_time_micro" json:"date_time_micro"`
		DateTimeNano   *carbon.DateTimeNano   `form:"date_time_nano" json:"date_time_nano"`
		Date           *carbon.Date           `form:"date" json:"date"`
		DateMilli      *carbon.DateMilli      `form:"date_milli" json:"date_milli"`
		DateMicro      *carbon.DateMicro      `form:"date_micro" json:"date_micro"`
		DateNano       *carbon.DateNano       `form:"date_nano" json:"date_nano"`
		Timestamp      *carbon.Timestamp      `form:"timestamp" json:"timestamp"`
		TimestampMilli *carbon.TimestampMilli `form:"timestamp_milli" json:"timestamp_milli"`
		TimestampMicro *carbon.TimestampMicro `form:"timestamp_micro" json:"timestamp_micro"`
		TimestampNano  *carbon.TimestampNano  `form:"timestamp_nano" json:"timestamp_nano"`
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
	type Data struct {
		A string            `form:"a" json:"a"`
		B int               `form:"b" json:"b"`
		C int8              `form:"c" json:"c"`
		D int16             `form:"d" json:"d"`
		E int32             `form:"e" json:"e"`
		F int64             `form:"f" json:"f"`
		G uint              `form:"g" json:"g"`
		H uint8             `form:"h" json:"h"`
		I uint16            `form:"i" json:"i"`
		J uint32            `form:"j" json:"j"`
		K uint64            `form:"k" json:"k"`
		L bool              `form:"l" json:"l"`
		M float32           `form:"m" json:"m"`
		N float64           `form:"n" json:"n"`
		O []string          `form:"o" json:"o"`
		P map[string]string `form:"p" json:"p"`
	}

	tests := []struct {
		name       string
		data       validate.DataFace
		rules      map[string]string
		filters    map[string]string
		expectData Data
		expectErr  error
	}{
		{
			name: "success without cast data",
			data: func() validate.DataFace {
				body := &Data{
					A: "1",
					B: 1,
					C: 1,
					D: 1,
					E: 1,
					F: 1,
					G: 1,
					H: 1,
					I: 1,
					J: 1,
					K: 1,
					L: true,
					M: 1,
					N: 1,
					O: []string{"1"},
					P: map[string]string{"a": "aa"},
				}
				jsonStr, err := json.NewJson().Marshal(body)
				assert.Nil(t, err)
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonStr))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{
				"a": "required",
				"b": "required",
				"c": "required",
				"d": "required",
				"e": "required",
				"f": "required",
				"g": "required",
				"h": "required",
				"i": "required",
				"j": "required",
				"k": "required",
				"l": "required",
				"m": "required",
				"n": "required",
				"o": "required",
				"p": "required",
			},
			filters: map[string]string{},
			expectData: Data{
				A: "1",
				B: 1,
				C: 1,
				D: 1,
				E: 1,
				F: 1,
				G: 1,
				H: 1,
				I: 1,
				J: 1,
				K: 1,
				L: true,
				M: 1,
				N: 1,
				O: []string{"1"},
				P: map[string]string{"a": "aa"},
			},
		}, {
			name: "success with cast data",
			data: func() validate.DataFace {
				body := map[string]any{
					"a": 1,
					"b": "1",
					"c": "1",
					"d": "1",
					"e": "1",
					"f": "1",
					"g": "1",
					"h": "1",
					"i": "1",
					"j": "1",
					"k": "1",
					"l": "true",
					"m": "1",
					"n": "1",
					"o": []int{1},
					"p": map[string]string{"a": "aa"},
				}
				jsonStr, err := json.NewJson().Marshal(body)
				assert.Nil(t, err)
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonStr))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{
				"a": "required",
				"b": "required",
				"c": "required",
				"d": "required",
				"e": "required",
				"f": "required",
				"g": "required",
				"h": "required",
				"i": "required",
				"j": "required",
				"k": "required",
				"l": "required",
				"m": "required",
				"n": "required",
				"o": "required",
				"p": "required",
			},
			filters: map[string]string{},
			expectData: Data{
				A: "1",
				B: 1,
				C: 1,
				D: 1,
				E: 1,
				F: 1,
				G: 1,
				H: 1,
				I: 1,
				J: 1,
				K: 1,
				L: true,
				M: 1,
				N: 1,
				O: []string{"1"},
				P: map[string]string{"a": "aa"},
			},
		},
	}

	validation := NewValidation()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.Make(test.data, test.rules, Filters(test.filters))
			assert.Nil(t, err)

			var data Data
			err = validator.Bind(&data)
			assert.Nil(t, test.expectErr, err)
			assert.Equal(t, test.expectData.A, data.A)
			assert.Equal(t, test.expectData.B, data.B)
			assert.Equal(t, test.expectData.C, data.C)
			assert.Equal(t, test.expectData.D, data.D)
			assert.Equal(t, test.expectData.E, data.E)
			assert.Equal(t, test.expectData.F, data.F)
			assert.Equal(t, test.expectData.G, data.G)
			assert.Equal(t, test.expectData.H, data.H)
			assert.Equal(t, test.expectData.I, data.I)
			assert.Equal(t, test.expectData.J, data.J)
			assert.Equal(t, test.expectData.K, data.K)
			assert.Equal(t, test.expectData.L, data.L)
			assert.Equal(t, test.expectData.M, data.M)
			assert.Equal(t, test.expectData.N, data.N)
			assert.Equal(t, test.expectData.O, data.O)
			assert.Equal(t, test.expectData.P, data.P)
		})
	}
}

func TestCastCarbon(t *testing.T) {
	tests := []struct {
		name      string
		from      reflect.Value
		transform func(carbon carbon.Carbon) any
		assert    func(result any)
	}{
		{
			name: "Happy path - length 10 string",
			from: reflect.ValueOf("2024-07-04"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewDate(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.Date{}, result)
				assert.Equal(t, "2024-07-04", result.(carbon.Date).ToDateString())
			},
		},
		{
			name: "Happy path - length 10 int",
			from: reflect.ValueOf(1720087252),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestamp(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.Timestamp{}, result)
				assert.Equal(t, "2024-07-04 10:00:52", result.(carbon.Timestamp).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 13 int",
			from: reflect.ValueOf(1720087252123),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(carbon.TimestampMilli).ToDateTimeMilliString())
			},
		},
		{
			name: "Sad path - length 13 string",
			from: reflect.ValueOf("1720087252123"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(carbon.TimestampMilli).ToDateTimeMilliString())
			},
		},
		{
			name: "Happy path - length 13 Y-m-d H",
			from: reflect.ValueOf("2024-07-04 10"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampMilli(c)
			},
			assert: func(result any) {
				assert.Equal(t, "2024-07-04 10:00:00", result.(carbon.TimestampMilli).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 16 int",
			from: reflect.ValueOf(1720087252123456),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampMicro(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampMicro{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456", result.(carbon.TimestampMicro).ToDateTimeMicroString())
			},
		},
		{
			name: "Happy path - length 16 string",
			from: reflect.ValueOf("1720087252123456"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampMicro(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampMicro{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456", result.(carbon.TimestampMicro).ToDateTimeMicroString())
			},
		},
		{
			name: "Happy path - length 16 Y-m-d H:i",
			from: reflect.ValueOf("2024-07-04 10:00"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewDateTime(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.DateTime{}, result)
				assert.Equal(t, "2024-07-04 10:00:00", result.(carbon.DateTime).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length 19 int",
			from: reflect.ValueOf(1720087252123456789),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampNano(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampNano{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456789", result.(carbon.TimestampNano).ToDateTimeNanoString())
			},
		},
		{
			name: "Happy path - length 19 int",
			from: reflect.ValueOf("1720087252123456789"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewTimestampNano(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.TimestampNano{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123456789", result.(carbon.TimestampNano).ToDateTimeNanoString())
			},
		},
		{
			name: "Happy path - length 19 Y-m-d H:i:s",
			from: reflect.ValueOf("2024-07-04 10:00:52"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewDateTime(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.DateTime{}, result)
				assert.Equal(t, "2024-07-04 10:00:52", result.(carbon.DateTime).ToDateTimeString())
			},
		},
		{
			name: "Happy path - length other",
			from: reflect.ValueOf("2024-07-04 10:00:52.123"),
			transform: func(c carbon.Carbon) any {
				return carbon.NewDateTimeMilli(c)
			},
			assert: func(result any) {
				assert.IsType(t, carbon.DateTimeMilli{}, result)
				assert.Equal(t, "2024-07-04 10:00:52.123", result.(carbon.DateTimeMilli).ToDateTimeMilliString())
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

	defer func() {
		_ = logo.Close()
	}()
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
