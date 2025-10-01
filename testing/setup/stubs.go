package main

type Stubs struct{}

func (s Stubs) ExampleTest() string {
	return `package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type ExampleTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *ExampleTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *ExampleTestSuite) TearDownTest() {
}

func (s *ExampleTestSuite) TestIndex() {
	s.True(true)
}
`
}

func (s Stubs) TestCase() string {
	return `package tests

import (
	"github.com/goravel/framework/testing"

	"goravel/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
`
}

func (s Stubs) TestingFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/testing"
)

func Testing() testing.Testing {
	return App().MakeTesting()
}
`
}
