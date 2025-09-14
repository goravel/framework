package console

type Stubs struct {
}

func (r Stubs) Test() string {
	return `package DummyPackage

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"DummyModule/tests"
)

type DummyTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestDummyTestSuite(t *testing.T) {
	suite.Run(t, new(DummyTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *DummyTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *DummyTestSuite) TearDownTest() {
}

func (s *DummyTestSuite) TestIndex() {
	// TODO
}
`
}

func (r Stubs) ServiceProvider() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

type DummyServiceProvider struct{}

// Relationship provides the service provider's bindings, their dependencies, and the services they provide for.
// It's optional if the service provider doesn't depend on or provide any other services.
func (r *DummyServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}
// Register service bindings here
func (r *DummyServiceProvider) Register(app foundation.Application) {
	// Example:
	// app.Singleton("example", func(app foundation.Application) (any, error) {
	//     return &ExampleService{}, nil
	// })
}

// Boot performs post-registration booting of services.
// It will be called after all service providers have been registered.
func (r *DummyServiceProvider) Boot(app foundation.Application) {
}
`
}
