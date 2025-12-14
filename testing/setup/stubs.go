package main

import "strings"

type Stubs struct{}

func (s Stubs) ExampleTest(testsImport, testsPackage string) string {
	content := `package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"DummyTestsImport"
)

type ExampleTestSuite struct {
	suite.Suite
	DummyTestsPackage.TestCase
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

	content = strings.ReplaceAll(content, "DummyTestsImport", testsImport)
	content = strings.ReplaceAll(content, "DummyTestsPackage", testsPackage)

	return content
}

func (s Stubs) TestCase(pkg, bootstrapImport, bootstrapPackage string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/testing"

	"DummyBootstrapImport"
)

func init() {
	DummyBootstrapPackage.Boot()
}

type TestCase struct {
	testing.TestCase
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyBootstrapImport", bootstrapImport)
	content = strings.ReplaceAll(content, "DummyBootstrapPackage", bootstrapPackage)

	return content
}

func (s Stubs) TestingFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/testing"
)

func Testing() testing.Testing {
	return App().MakeTesting()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
