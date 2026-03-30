package foundation

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"
)

type ContainerTestSuite struct {
	suite.Suite
	container *Container
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}

func (s *ContainerTestSuite) SetupTest() {
	s.container = NewContainer()
}

func (s *ContainerTestSuite) TestBind() {
	callback := func(app foundation.Application) (any, error) {
		return 1, nil
	}
	s.container.Bind("Bind", callback)

	concrete, exist := s.container.bindings.Load("Bind")
	s.True(exist)
	ins, ok := concrete.(instance)
	s.True(ok)
	s.False(ins.shared)
	s.NotNil(ins.concrete)
	switch concrete := ins.concrete.(type) {
	case func(app foundation.Application) (any, error):
		concreteImpl, err := concrete(nil)
		s.Equal(1, concreteImpl)
		s.Nil(err)
	default:
		s.T().Errorf("error")
	}
}

func (s *ContainerTestSuite) TestBindings() {
	callback := func(app foundation.Application) (any, error) {
		return 1, nil
	}
	s.container.Bind("Bind", callback)

	s.ElementsMatch([]any{"Bind"}, s.container.Bindings())
}

func (s *ContainerTestSuite) TestBindWith() {
	callback := func(app foundation.Application, parameters map[string]any) (any, error) {
		return parameters["name"], nil
	}
	s.container.BindWith("BindWith", callback)

	concrete, exist := s.container.bindings.Load("BindWith")
	s.True(exist)
	ins, ok := concrete.(instance)
	s.True(ok)
	s.False(ins.shared)
	s.NotNil(ins.concrete)
	switch concrete := ins.concrete.(type) {
	case func(app foundation.Application, parameters map[string]any) (any, error):
		concreteImpl, err := concrete(nil, map[string]any{"name": "goravel"})
		s.Equal("goravel", concreteImpl)
		s.Nil(err)
	default:
		s.T().Errorf("error")
	}
}

func (s *ContainerTestSuite) TestInstance() {
	impl := 1
	s.container.Instance("Instance", impl)

	concrete, exist := s.container.bindings.Load("Instance")
	s.True(exist)
	ins, ok := concrete.(instance)
	s.True(ok)
	s.True(ins.shared)
	s.NotNil(ins.concrete)
	s.Equal(impl, ins.concrete)
}

func (s *ContainerTestSuite) TestSingleton_Fresh() {
	callback := func(app foundation.Application) (any, error) {
		return 1, nil
	}
	s.container.Singleton(binding.Config, callback)
	s.container.Singleton("Singleton", callback)

	res, err := s.container.Make(binding.Config)
	s.Nil(err)
	s.Equal(1, res)

	res, err = s.container.Make("Singleton")
	s.Nil(err)
	s.Equal(1, res)

	ins, ok := s.container.instances.Load("Singleton")
	s.True(ok)
	s.Equal(1, ins)

	s.container.Fresh("Singleton")

	res, ok = s.container.instances.Load("Singleton")
	s.False(ok)
	s.Nil(res)

	res, ok = s.container.instances.Load(binding.Config)
	s.True(ok)
	s.Equal(1, res)

	res, err = s.container.Make("Singleton")
	s.Nil(err)
	s.Equal(1, res)

	ins, ok = s.container.instances.Load("Singleton")
	s.True(ok)
	s.Equal(1, ins)

	s.container.Fresh()

	res, ok = s.container.instances.Load("Singleton")
	s.False(ok)
	s.Nil(res)

	res, ok = s.container.instances.Load(binding.Config)
	s.True(ok)
	s.Equal(1, res)
}

func (s *ContainerTestSuite) TestMake() {
	tests := []struct {
		name       string
		key        string
		parameters map[string]any
		setup      func()
		expectImpl any
		expectErr  error
	}{
		{
			name:      "not found Binding",
			key:       "no",
			setup:     func() {},
			expectErr: NewBindingNotFoundError("no"),
		},
		{
			name: "found Singleton",
			key:  "Singleton",
			setup: func() {
				s.container.Singleton("Singleton", func(app foundation.Application) (any, error) {
					return 1, nil
				})
			},
			expectImpl: 1,
		},
		{
			name: "found Bind",
			key:  "Bind",
			setup: func() {
				s.container.Bind("Bind", func(app foundation.Application) (any, error) {
					return 1, nil
				})
			},
			expectImpl: 1,
		},
		{
			name: "found BindWith",
			key:  "BindWith",
			parameters: map[string]any{
				"name": "goravel",
			},
			setup: func() {
				s.container.BindWith("BindWith", func(app foundation.Application, parameters map[string]any) (any, error) {
					return parameters["name"], nil
				})
			},
			expectImpl: "goravel",
		},
		{
			name: "found Instance",
			key:  "Instance",
			parameters: map[string]any{
				"name": "goravel",
			},
			setup: func() {
				s.container.Instance("Instance", 1)
			},
			expectImpl: 1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			impl, err := s.container.make(test.key, test.parameters)
			s.Equal(test.expectImpl, impl)
			s.Equal(test.expectErr, err)
		})
	}
}

func (s *ContainerTestSuite) TestMakeWrappers_SuppressBindingNotFoundError() {
	tests := []struct {
		name string
		run  func(container *Container) any
	}{
		{name: "ai", run: func(container *Container) any { return container.MakeAI() }},
		{name: "artisan", run: func(container *Container) any { return container.MakeArtisan() }},
		{name: "auth", run: func(container *Container) any { return container.MakeAuth() }},
		{name: "cache", run: func(container *Container) any { return container.MakeCache() }},
		{name: "config", run: func(container *Container) any { return container.MakeConfig() }},
		{name: "crypt", run: func(container *Container) any { return container.MakeCrypt() }},
		{name: "db", run: func(container *Container) any { return container.MakeDB() }},
		{name: "event", run: func(container *Container) any { return container.MakeEvent() }},
		{name: "gate", run: func(container *Container) any { return container.MakeGate() }},
		{name: "grpc", run: func(container *Container) any { return container.MakeGrpc() }},
		{name: "hash", run: func(container *Container) any { return container.MakeHash() }},
		{name: "http", run: func(container *Container) any { return container.MakeHttp() }},
		{name: "lang", run: func(container *Container) any { return container.MakeLang(context.Background()) }},
		{name: "log", run: func(container *Container) any { return container.MakeLog() }},
		{name: "mail", run: func(container *Container) any { return container.MakeMail() }},
		{name: "orm", run: func(container *Container) any { return container.MakeOrm() }},
		{name: "process", run: func(container *Container) any { return container.MakeProcess() }},
		{name: "queue", run: func(container *Container) any { return container.MakeQueue() }},
		{name: "rate_limiter", run: func(container *Container) any { return container.MakeRateLimiter() }},
		{name: "route", run: func(container *Container) any { return container.MakeRoute() }},
		{name: "schedule", run: func(container *Container) any { return container.MakeSchedule() }},
		{name: "schema", run: func(container *Container) any { return container.MakeSchema() }},
		{name: "seeder", run: func(container *Container) any { return container.MakeSeeder() }},
		{name: "session", run: func(container *Container) any { return container.MakeSession() }},
		{name: "storage", run: func(container *Container) any { return container.MakeStorage() }},
		{name: "telemetry", run: func(container *Container) any { return container.MakeTelemetry() }},
		{name: "testing", run: func(container *Container) any { return container.MakeTesting() }},
		{name: "validation", run: func(container *Container) any { return container.MakeValidation() }},
		{name: "view", run: func(container *Container) any { return container.MakeView() }},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			container := NewContainer()
			output := color.CaptureOutput(func(_ io.Writer) {
				s.Nil(test.run(container))
			})

			s.Equal("", output)
		})
	}
}

func (s *ContainerTestSuite) TestMakeWrappers_LogNonBindingError() {
	tests := []struct {
		name      string
		setup     func(container *Container) error
		run       func(container *Container) any
		expectErr string
	}{
		{
			name: "make callback returns error",
			setup: func(container *Container) error {
				expectedErr := fmt.Errorf("make callback error")
				container.Bind(facades.FacadeToBinding[facades.AI], func(app foundation.Application) (any, error) {
					return nil, expectedErr
				})

				return expectedErr
			},
			run:       func(container *Container) any { return container.MakeAI() },
			expectErr: "make callback error",
		},
		{
			name: "make with callback returns error",
			setup: func(container *Container) error {
				expectedErr := fmt.Errorf("make with callback error")
				container.BindWith(facades.FacadeToBinding[facades.Auth], func(app foundation.Application, parameters map[string]any) (any, error) {
					return nil, expectedErr
				})

				return expectedErr
			},
			run:       func(container *Container) any { return container.MakeAuth() },
			expectErr: "make with callback error",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			container := NewContainer()
			expectedErr := test.setup(container)
			output := color.CaptureOutput(func(_ io.Writer) {
				s.Nil(test.run(container))
			})

			s.Contains(output, expectedErr.Error())
			s.Contains(output, test.expectErr)
		})
	}
}
