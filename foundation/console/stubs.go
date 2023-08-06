package console

type Stubs struct {
}

func (r Stubs) Test() string {
	return `package DummyPackage

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
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
