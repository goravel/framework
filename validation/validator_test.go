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

	"github.com/goravel/framework/support/json"
)

func TestBind(t *testing.T) {
	type Data struct {
		A    string
		B    int
		C    string
		D    *Data
		File *multipart.FileHeader
	}

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
			name:  "success when data is map and key is int",
			data:  validate.FromMap(map[string]any{"b": 1}),
			rules: map[string]string{"b": "required"},
			expectData: Data{
				B: 1,
			},
		},
		{
			name:  "success when data is map and cast key",
			data:  validate.FromMap(map[string]any{"b": "1"}),
			rules: map[string]string{"b": "required"},
			expectData: Data{
				B: 1,
			},
		},
		{
			name:  "success when data is map, key is lowercase and has errors",
			data:  validate.FromMap(map[string]any{"a": "aa", "c": "cc"}),
			rules: map[string]string{"a": "required", "b": "required"},
			expectData: Data{
				A: "",
				C: "",
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
			name: "empty when data is struct and key is struct",
			data: func() validate.DataFace {
				data, err := validate.FromStruct(struct {
					D *Data
				}{
					D: &Data{
						A: "aa",
					},
				})
				assert.Nil(t, err)

				return data
			}(),
			rules:      map[string]string{"d.a": "required"},
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
			rules: map[string]string{"a": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "success when data is get request and params is int",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodGet, "/?b=1", nil)
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"b": "required"},
			expectData: Data{
				B: 1,
			},
		},
		{
			name: "success when data is post request",
			data: func() validate.DataFace {
				request, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"a":"aa"}`))
				request.Header.Set("Content-Type", "application/json")
				assert.Nil(t, err)
				data, err := validate.FromRequest(request)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			name: "success when data is post request with body",
			data: func() validate.DataFace {
				request := buildRequest(t)
				data, err := validate.FromRequest(request, 1)
				assert.Nil(t, err)

				return data
			}(),
			rules: map[string]string{"a": "required", "file": "required"},
			expectData: func() Data {
				request := buildRequest(t)
				_, fileHeader, _ := request.FormFile("file")
				data := Data{
					A:    "aa",
					File: fileHeader,
				}

				return data
			}(),
		},
	}

	validation := NewValidation()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, err := validation.Make(test.data, test.rules)
			assert.Nil(t, err)

			var data Data
			err = validator.Bind(&data)
			assert.Nil(t, test.expectErr, err)
			assert.Equal(t, test.expectData.A, data.A)
			assert.Equal(t, test.expectData.B, data.B)
			assert.Equal(t, test.expectData.C, data.C)
			assert.Equal(t, test.expectData.D, data.D)
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
				jsonStr, err := json.Marshal(body)
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
				jsonStr, err := json.Marshal(body)
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
			validator, err := validation.Make(test.data, test.rules)
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
