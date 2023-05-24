package foundation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
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
	callback := func() (any, error) {
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
	case func() (any, error):
		concreteImpl, err := concrete()
		s.Equal(1, concreteImpl)
		s.Nil(err)
	default:
		s.T().Errorf("error")
	}
}

func (s *ContainerTestSuite) TestBindWith() {
	callback := func(parameters map[string]any) (any, error) {
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
	case func(parameters map[string]any) (any, error):
		concreteImpl, err := concrete(map[string]any{"name": "goravel"})
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

func (s *ContainerTestSuite) TestSingleton() {
	callback := func() (any, error) {
		return 1, nil
	}
	s.container.Singleton("Singleton", callback)

	concrete, exist := s.container.bindings.Load("Singleton")
	s.True(exist)
	ins, ok := concrete.(instance)
	s.True(ok)
	s.True(ins.shared)
	s.NotNil(ins.concrete)
	switch concrete := ins.concrete.(type) {
	case func() (any, error):
		concreteImpl, err := concrete()
		s.Equal(1, concreteImpl)
		s.Nil(err)
	default:
		s.T().Errorf("error")
	}
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
			name:      "not found binding",
			key:       "no",
			setup:     func() {},
			expectErr: fmt.Errorf("binding not found: %+v", "no"),
		},
		{
			name: "found Singleton",
			key:  "Singleton",
			setup: func() {
				s.container.Singleton("Singleton", func() (any, error) {
					return 1, nil
				})
			},
			expectImpl: 1,
		},
		{
			name: "found Bind",
			key:  "Bind",
			setup: func() {
				s.container.Bind("Bind", func() (any, error) {
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
				s.container.BindWith("BindWith", func(parameters map[string]any) (any, error) {
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
