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

type ResponseFactoryTestSuite struct {
	suite.Suite
	factory *ResponseFactory
}

func TestResponseFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseFactoryTestSuite))
}

func (s *ResponseFactoryTestSuite) SetupTest() {
	s.factory = NewResponseFactory(json.New())
}

func (s *ResponseFactoryTestSuite) TestJson() {
	s.Run("Success", func() {
		data := map[string]any{
			"name": "Goravel",
			"meta": map[string]int{"id": 1},
		}

		response := s.factory.Json(data, http.StatusCreated)
		s.Equal(http.StatusCreated, response.Status())
		s.Equal("application/json", response.Header("Content-Type"))

		body, err := response.Json()
		s.NoError(err)
		s.Equal("Goravel", body["name"])
		s.Equal(map[string]any{"id": float64(1)}, body["meta"])
	})

	s.Run("Marshal Error", func() {
		invalidData := make(chan int) // Channels cannot be marshaled
		response := s.factory.Json(invalidData, http.StatusOK)

		s.Equal(http.StatusInternalServerError, response.Status())

		bodyStr, err := response.Body()
		s.NoError(err)
		s.Contains(bodyStr, "json: unsupported type")
	})
}

func (s *ResponseFactoryTestSuite) TestBasicResponses() {
	tests := []struct {
		name           string
		response       client.Response
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "String Response",
			response:       s.factory.String("Hello World", http.StatusNotFound),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Hello World",
		},
		{
			name:           "Status Only (Teapot)",
			response:       s.factory.Status(http.StatusTeapot),
			expectedStatus: http.StatusTeapot,
			expectedBody:   "",
		},
		{
			name:           "Success Helper",
			response:       s.factory.Success(),
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

func (s *ResponseFactoryTestSuite) TestFile() {
	s.Run("Success", func() {
		dir := s.T().TempDir()
		filePath := filepath.Join(dir, "test_file.txt")
		content := "File Content"

		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		response := s.factory.File(filePath, http.StatusOK)

		s.Equal(http.StatusOK, response.Status())
		body, err := response.Body()
		s.NoError(err)
		s.Equal(content, body)
	})

	s.Run("Not Found", func() {
		response := s.factory.File("non_existent_file.txt", http.StatusOK)

		s.Equal(http.StatusInternalServerError, response.Status())
		body, err := response.Body()
		s.NoError(err)
		s.Contains(body, "File not found")
	})
}

func (s *ResponseFactoryTestSuite) TestMake_Headers() {
	headers := map[string]string{
		"X-Custom-Header": "Goravel",
		"Cache-Control":   "no-cache",
	}

	response := s.factory.Make("body", http.StatusOK, headers)

	s.Equal("Goravel", response.Header("X-Custom-Header"))
	s.Equal("no-cache", response.Header("Cache-Control"))
}
