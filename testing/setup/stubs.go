package main

import "strings"

type Stubs struct{}

func (s Stubs) ExampleTest(imt, testsPackage string) string {
	content := `package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"DummyImport"
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

	content = strings.ReplaceAll(content, "DummyImport", imt)
	content = strings.ReplaceAll(content, "DummyTestsPackage", testsPackage)

	return content
}

func (s Stubs) TestCase(pkg, bootstrap string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/testing"

	"DummyBootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyBootstrap", bootstrap)

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
