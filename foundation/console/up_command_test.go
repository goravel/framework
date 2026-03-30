package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

type UpCommandTestSuite struct {
	suite.Suite
	mockStorage *mocksfilesystem.Storage
}

func TestUpCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UpCommandTestSuite))
}

func (s *UpCommandTestSuite) SetupSuite() {
	s.mockStorage = mocksfilesystem.NewStorage(s.T())
}

func (s *UpCommandTestSuite) TearDownSuite() {
}

func (s *UpCommandTestSuite) TestSignature() {
	expected := "up"
	s.Require().Equal(expected, NewUpCommand(s.mockStorage).Signature())
}

func (s *UpCommandTestSuite) TestDescription() {
	expected := "Bring the application out of maintenance mode"
	s.Require().Equal(expected, NewUpCommand(s.mockStorage).Description())
}

func (s *UpCommandTestSuite) TestExtend() {
	cmd := NewUpCommand(s.mockStorage)
	got := cmd.Extend()

	s.Empty(got)
}

func (s *UpCommandTestSuite) TestHandle() {
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(true).Once()
	s.mockStorage.EXPECT().Delete("framework/maintenance.json").Return(nil).Once()

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Success("The application is up and live now").Once()

	cmd := NewUpCommand(s.mockStorage)
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}

func (s *UpCommandTestSuite) TestHandleWhenNotDown() {
	s.mockStorage.EXPECT().Exists("framework/maintenance.json").Return(false).Once()
	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Error("The application is not in maintenance mode").Once()

	cmd := NewUpCommand(s.mockStorage)
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}
