package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractshttp "github.com/goravel/framework/contracts/http"
	mockhttp "github.com/goravel/framework/mocks/http"
	mocksession "github.com/goravel/framework/mocks/session"
)

func TestTokenMatch(t *testing.T) {
	tests := []struct {
		name          string
		hasSession    bool
		sessionToken  string
		headerToken   string
		formToken     string
		expectedMatch bool
	}{
		{
			name:          "no session returns false",
			hasSession:    false,
			expectedMatch: false,
		},
		{
			name:          "valid token in header",
			hasSession:    true,
			sessionToken:  "valid-token",
			headerToken:   "valid-token",
			expectedMatch: true,
		},
		{
			name:          "valid token in form",
			hasSession:    true,
			sessionToken:  "valid-token",
			formToken:     "valid-token",
			expectedMatch: true,
		},
		{
			name:          "invalid token",
			hasSession:    true,
			sessionToken:  "valid-token",
			headerToken:   "invalid-token",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := mockhttp.NewContext(t)
			mockRequest := mockhttp.NewContextRequest(t)
			mockSession := mocksession.NewSession(t)

			mockRequest.EXPECT().HasSession().Return(tt.hasSession).Once()

			if tt.hasSession {
				mockRequest.EXPECT().Session().Return(mockSession).Once()
				mockSession.EXPECT().Token().Return(tt.sessionToken).Once()
				mockRequest.EXPECT().Header(HeaderCsrfKey).Return(tt.headerToken).Once()
				if tt.headerToken == "" {
					mockCtx.EXPECT().Request().Return(mockRequest).Times(4)
					mockRequest.EXPECT().Input("_token").Return(tt.formToken).Once()
				} else {
					mockCtx.EXPECT().Request().Return(mockRequest).Times(3)
				}
			} else {
				mockCtx.EXPECT().Request().Return(mockRequest).Once()
			}
			result := tokenMatch(mockCtx)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}

func TestInExceptArray(t *testing.T) {
	tests := []struct {
		name        string
		excepts     []string
		currentPath string
		expected    bool
	}{
		{
			name:        "exact match",
			excepts:     []string{"api/users"},
			currentPath: "api/users",
			expected:    true,
		},
		{
			name:        "wildcard match",
			excepts:     []string{"api/*"},
			currentPath: "api/users",
			expected:    true,
		},
		{
			name:        "no match",
			excepts:     []string{"api/*"},
			currentPath: "web/users",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inExceptArray(tt.excepts, tt.currentPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsReading(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"GET method", contractshttp.MethodGet, true},
		{"HEAD method", contractshttp.MethodHead, true},
		{"OPTIONS method", contractshttp.MethodOptions, true},
		{"POST method", contractshttp.MethodPost, false},
		{"PUT method", contractshttp.MethodPut, false},
		{"DELETE method", contractshttp.MethodDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReading(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseExceptPaths(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string
		expected []string
	}{
		{
			name:     "simple paths",
			inputs:   []string{"/api/users", "web/posts/"},
			expected: []string{"api/users", "web/posts"},
		},
		{
			name:     "with query parameters",
			inputs:   []string{"/api/users?page=1", "web/posts?sort=desc"},
			expected: []string{"api/users", "web/posts"},
		},
		{
			name:     "with wildcards",
			inputs:   []string{"/api/*", "web/*/comments"},
			expected: []string{"api/*", "web/*/comments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExceptPaths(tt.inputs)
			assert.Equal(t, tt.expected, result)
		})
	}
}
