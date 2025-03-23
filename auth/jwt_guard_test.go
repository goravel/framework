package auth

import (
	"testing"

	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
)

func TestGetTtl(t *testing.T) {
	var mockConfig *mocksconfig.Config

	tests := []struct {
		name     string
		setup    func()
		expected int
	}{
		{
			name: "GuardTtlIsNil",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				mockConfig.EXPECT().GetInt("jwt.ttl").Return(2).Once()
			},
			expected: 2,
		},
		{
			name: "GuardTtlIsNotNil",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(1).Once()
			},
			expected: 1,
		},
		{
			name: "GuardTtlIsZero",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
		{
			name: "JwtTtlIsZero",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				mockConfig.EXPECT().GetInt("jwt.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)

			test.setup()

			ttl := getTtl(mockConfig, testUserGuard)
			assert.Equal(t, test.expected, ttl)
		})
	}
}
