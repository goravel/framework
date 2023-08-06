package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/validate"
	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	type Data struct {
		A    string
		B    int
		C    string
		File *multipart.FileHeader
	}

	request := buildRequest(t)

	tests := []struct {
		name       string
		data       validate.DataFace
		rules      map[string]string
		expectData Data
		expectErr  error
	}{
		{
			name:  "success when data is map and key is lowercase",
			data:  validate.FromMap(map[string]any{"a": "aa"}),
			rules: map[string]string{"a": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name:  "success when data is map, key is lowercase and has errors",
			data:  validate.FromMap(map[string]any{"a": "aa", "c": "cc"}),
			rules: map[string]string{"a": "required", "b": "required"},
			expectData: Data{
				A: "aa",
				C: "cc",
			},
		},
		{
			name:  "success when data is map and key is uppercase",
			data:  validate.FromMap(map[string]any{"A": "aa"}),
			rules: map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "success when data is struct and key is uppercase",
			data: func() validate.DataFace {
				data, err := validate.FromStruct(struct {
					A string
				}{
					A: "aa",
				})
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "empty when data is struct and key is lowercase",
			data: func() validate.DataFace {
				data, err := validate.FromStruct(struct {
					a string
				}{
					a: "aa",
				})
				assert.Nil(t, err)

				return data
			}(),
			rules:      map[string]string{"a": "required"},
			expectData: Data{},
		},
		{
			name: "success when data is get request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodGet, "/?a=aa", nil)
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "success when data is post request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodGet, "/?a=aa", nil)
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "success when data is post request with body",
			data: func() validate.DataFace {
				data, err := validate.FromRequest(request, 1)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"A": "required", "File": "required"},
			expectData: func() Data {
				_, fileHeader, _ := request.FormFile("file")
				data := Data{
					A:    "aa",
					File: fileHeader,
				}

				return data
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator := &Validator{data: test.data}

			var data Data
			err := validator.Bind(&data)
			assert.Nil(t, test.expectErr, err)
			assert.Equal(t, test.expectData.A, data.A)
			assert.Equal(t, test.expectData.B, data.B)
			assert.Equal(t, test.expectData.C, data.C)
			assert.Equal(t, test.expectData.File == nil, data.File == nil)
		})
	}
}

func TestFails(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		expectRes bool
	}{
		{
			describe: "false",
			data:     map[string]any{"a": "aa"},
			rules:    map[string]string{"a": "required"},
		},
		{
			describe:  "true",
			data:      map[string]any{"b": "bb"},
			rules:     map[string]string{"a": "required"},
			expectRes: true,
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			test.data,
			test.rules,
		)
		assert.Nil(t, err)
		assert.Equal(t, test.expectRes, validator.Fails(), test.describe)
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
