package client

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

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

func (s *FakeResponseTestSuite) TestFile() {
	dir := s.T().TempDir()
	validFile := filepath.Join(dir, "test.txt")
	s.Require().NoError(os.WriteFile(validFile, []byte("Goravel File Content"), 0644))

	tests := []struct {
		name           string
		path           string
		status         int
		expectedStatus int
		expectedBody   string
		expectError    bool
	}{
		{
			name:           "Success reads file content",
			path:           validFile,
			status:         http.StatusOK,
			expectedStatus: http.StatusOK,
			expectedBody:   "Goravel File Content",
		},
		{
			name:           "Error when file does not exist",
			path:           filepath.Join(dir, "missing.txt"),
			status:         http.StatusOK,
			expectedStatus: http.StatusInternalServerError,
			// Matches the exact error message in implementation
			expectedBody: "Failed to read mock file",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := s.fakeResponse.File(tt.path, tt.status)

			s.Equal(tt.expectedStatus, resp.Status())

			body, err := resp.Body()
			s.NoError(err)
			if tt.expectError {
				s.Contains(body, tt.expectedBody)
				return
			}

			s.Equal(tt.expectedBody, body)
		})
	}
}

func (s *FakeResponseTestSuite) TestJson() {
	tests := []struct {
		name           string
		data           any
		status         int
		expectedStatus int
		expectedBody   string
		expectHeader   string
	}{
		{
			name:           "Success marshals map",
			data:           map[string]string{"foo": "bar"},
			status:         http.StatusCreated,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"foo":"bar"}`,
			expectHeader:   "application/json",
		},
		{
			name:           "Success marshals struct",
			data:           struct{ ID int }{ID: 1},
			status:         http.StatusOK,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"ID":1}`,
			expectHeader:   "application/json",
		},
		{
			name:           "Error marshals invalid channel",
			data:           make(chan int),
			status:         http.StatusOK,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to marshal mock JSON",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := s.fakeResponse.Json(tt.data, tt.status)

			s.Equal(tt.expectedStatus, resp.Status())

			if tt.expectHeader != "" {
				s.Equal(tt.expectHeader, resp.Header("Content-Type"))
			}

			body, err := resp.Body()
			s.NoError(err)
			if tt.expectedStatus == http.StatusInternalServerError {
				s.Contains(body, tt.expectedBody)
				return
			}

			s.JSONEq(tt.expectedBody, body)
		})
	}
}

func (s *FakeResponseTestSuite) TestMake() {
	tests := []struct {
		name            string
		body            string
		status          int
		headers         http.Header
		expectedStatus  int
		expectedBody    string
		expectedHeaders http.Header
	}{
		{
			name:           "Success with simple headers",
			body:           "Custom Body",
			status:         http.StatusTeapot,
			headers:        http.Header{"X-Test": []string{"True"}},
			expectedStatus: http.StatusTeapot,
			expectedBody:   "Custom Body",
			expectedHeaders: http.Header{
				"X-Test": []string{"True"},
			},
		},
		{
			name:           "Success with multi-value headers (Cookies)",
			body:           "Cookie Monster",
			status:         200,
			headers:        http.Header{"Set-Cookie": []string{"a=1", "b=2"}},
			expectedStatus: 200,
			expectedBody:   "Cookie Monster",
			expectedHeaders: http.Header{
				"Set-Cookie": []string{"a=1", "b=2"},
			},
		},
		{
			name:            "Success with empty headers",
			body:            "",
			status:          http.StatusOK,
			headers:         nil,
			expectedStatus:  http.StatusOK,
			expectedBody:    "",
			expectedHeaders: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := s.fakeResponse.Make(tt.body, tt.status, tt.headers)

			s.Equal(tt.expectedStatus, resp.Status())

			body, err := resp.Body()
			s.NoError(err)
			s.Equal(tt.expectedBody, body)

			for k, expectedVals := range tt.expectedHeaders {
				actualVals := resp.Headers().Values(k)
				s.ElementsMatch(expectedVals, actualVals)
			}
		})
	}
}

func (s *FakeResponseTestSuite) TestOK() {
	resp := s.fakeResponse.OK()
	s.Equal(http.StatusOK, resp.Status())

	body, err := resp.Body()
	s.NoError(err)
	s.Empty(body)
}

func (s *FakeResponseTestSuite) TestStatus() {
	tests := []struct {
		code int
	}{
		{http.StatusOK},
		{http.StatusBadRequest},
		{http.StatusInternalServerError},
	}

	for _, tt := range tests {
		s.Run("Status matches", func() {
			resp := s.fakeResponse.Status(tt.code)
			s.Equal(tt.code, resp.Status())
			body, err := resp.Body()
			s.NoError(err)
			s.Empty(body)
		})
	}
}

func (s *FakeResponseTestSuite) TestString() {
	tests := []struct {
		name           string
		body           string
		status         int
		expectedStatus int
	}{
		{
			name:           "Simple string",
			body:           "Hello",
			status:         http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty string",
			body:           "",
			status:         http.StatusNoContent,
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := s.fakeResponse.String(tt.body, tt.status)
			s.Equal(tt.expectedStatus, resp.Status())
			body, err := resp.Body()
			s.NoError(err)
			s.Equal(tt.body, body)
		})
	}
}
