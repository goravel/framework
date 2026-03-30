package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type ServiceProviderTestSuite struct {
	suite.Suite
}

func TestServiceProviderTestSuite(t *testing.T) {
	suite.Run(t, &ServiceProviderTestSuite{})
}

func (s *ServiceProviderTestSuite) TestRelationship() {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()
	s.Equal(binding.Relationship{
		Bindings:     []string{binding.AI},
		Dependencies: binding.Bindings[binding.AI].Dependencies,
	}, relationship)
}

func (s *ServiceProviderTestSuite) TestRegister() {
	var (
		mockApp         *mocksfoundation.Application
		mockCallbackApp *mocksfoundation.Application
		mockConfig      *mocksconfig.Config
		callback        func(contractsfoundation.Application) (any, error)
	)

	beforeEach := func() {
		mockApp = mocksfoundation.NewApplication(s.T())
		mockCallbackApp = mocksfoundation.NewApplication(s.T())
		mockConfig = mocksconfig.NewConfig(s.T())

		provider := &ServiceProvider{}
		mockApp.EXPECT().Singleton(binding.AI, mock.MatchedBy(func(cb any) bool {
			typedCallback, ok := cb.(func(contractsfoundation.Application) (any, error))
			if !ok {
				return false
			}
			callback = typedCallback
			return true
		})).Once()
		provider.Register(mockApp)
		s.Require().NotNil(callback)
	}

	tests := []struct {
		name        string
		setup       func()
		expectError error
	}{
		{
			name: "binds application",
			setup: func() {
				mockCallbackApp.EXPECT().MakeConfig().Return(mockConfig).Once()
				mockConfig.EXPECT().
					UnmarshalKey("ai", mock.MatchedBy(func(rawVal any) bool {
						_, ok := rawVal.(*contractsai.Config)
						return ok
					})).
					RunAndReturn(func(_ string, rawVal any) error {
						*rawVal.(*contractsai.Config) = contractsai.Config{Default: "default"}
						return nil
					}).
					Once()
			},
		},
		{
			name: "returns error when config cannot be unmarshaled",
			setup: func() {
				mockCallbackApp.EXPECT().MakeConfig().Return(mockConfig).Once()
				mockConfig.EXPECT().
					UnmarshalKey("ai", mock.MatchedBy(func(rawVal any) bool {
						_, ok := rawVal.(*contractsai.Config)
						return ok
					})).
					Return(assert.AnError).
					Once()
			},
			expectError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			beforeEach()
			tt.setup()

			instance, err := callback(mockCallbackApp)
			s.Equal(tt.expectError, err)
			if tt.expectError != nil {
				s.Nil(instance)
				return
			}
			s.IsType(&Application{}, instance)
			s.Equal(contractsai.Config{Default: "default"}, instance.(*Application).config)
		})
	}
}

func (s *ServiceProviderTestSuite) TestBoot() {
	provider := &ServiceProvider{}
	mockApp := mocksfoundation.NewApplication(s.T())

	mockApp.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 1 &&
			commands[0] != nil &&
			commands[0].Signature() == "make:agent"
	})).Once()

	provider.Boot(mockApp)
}
