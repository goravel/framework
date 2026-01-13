package client

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/foundation/json"
)

type FakeResponseTestSuite struct {
	suite.Suite
	fakeResponse *FakeResponse
}

func TestFakeResponseTestSuite(t *testing.T) {
	suite.Run(t, new(FakeResponseTestSuite))
}

func (s *FakeResponseTestSuite) SetupTest() {
	s.fakeResponse = NewFakeResponse(json.New())
}

func (s *FakeResponseTestSuite) TestJson() {
	s.Run("Success", func() {
		data := map[string]any{
			"name": "Goravel",
			"meta": map[string]int{"id": 1},
		}

		response := s.fakeResponse.Json(data, http.StatusCreated)
		s.Equal(http.StatusCreated, response.Status())
		s.Equal("application/json", response.Header("Content-Type"))

		body, err := response.Json()
		s.NoError(err)
		s.Equal("Goravel", body["name"])
		s.Equal(map[string]any{"id": float64(1)}, body["meta"])
	})

	s.Run("Marshal Error", func() {
		invalidData := make(chan int) // Channels cannot be marshaled
		response := s.fakeResponse.Json(invalidData, http.StatusOK)

		s.Equal(http.StatusInternalServerError, response.Status())

		bodyStr, err := response.Body()
		s.NoError(err)
		s.Contains(bodyStr, "json: unsupported type")
	})
}

func (s *FakeResponseTestSuite) TestBasicResponses() {
	tests := []struct {
		name           string
		response       client.Response
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "String Response",
			response:       s.fakeResponse.String("Hello World", http.StatusNotFound),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Hello World",
		},
		{
			name:           "Status Only (Teapot)",
			response:       s.fakeResponse.Status(http.StatusTeapot),
			expectedStatus: http.StatusTeapot,
			expectedBody:   "",
		},
		{
			name:           "OK Helper",
			response:       s.fakeResponse.OK(),
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expectedStatus, tt.response.Status())

			body, err := tt.response.Body()
			s.NoError(err)
			s.Equal(tt.expectedBody, body)
		})
	}
}

func (s *FakeResponseTestSuite) TestFile() {
	s.Run("Success", func() {
		dir := s.T().TempDir()
		filePath := filepath.Join(dir, "test_file.txt")
		content := "File Content"

		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		response := s.fakeResponse.File(filePath, http.StatusOK)

		s.Equal(http.StatusOK, response.Status())
		body, err := response.Body()
		s.NoError(err)
		s.Equal(content, body)
	})

	s.Run("Not Found", func() {
		response := s.fakeResponse.File("non_existent_file.txt", http.StatusOK)

		s.Equal(http.StatusInternalServerError, response.Status())
		body, err := response.Body()
		s.NoError(err)
		s.Contains(body, "File not found")
	})
}

func (s *FakeResponseTestSuite) TestMake_Headers() {
	headers := map[string]string{
		"X-Custom-Header": "Goravel",
		"Cache-Control":   "no-cache",
	}

	response := s.fakeResponse.Make("body", http.StatusOK, headers)

	s.Equal("Goravel", response.Header("X-Custom-Header"))
	s.Equal("no-cache", response.Header("Cache-Control"))
}
