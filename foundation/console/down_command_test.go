package console

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/file"
)

type DownCommandTestSuite struct {
	suite.Suite
}

func TestDownCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DownCommandTestSuite))
}

func (s *DownCommandTestSuite) SetupSuite() {
}

func (s *DownCommandTestSuite) TearDownSuite() {
}

func (s *DownCommandTestSuite) TestSignature() {
	expected := "down"
	s.Require().Equal(expected, NewDownCommand(mocksfoundation.NewApplication(s.T())).Signature())
}

func (s *DownCommandTestSuite) TestDescription() {
	expected := "Put the application into maintenance mode"
	s.Require().Equal(expected, NewDownCommand(mocksfoundation.NewApplication(s.T())).Description())
}

func (s *DownCommandTestSuite) TestExtend() {
	cmd := NewDownCommand(mocksfoundation.NewApplication(s.T()))
	got := cmd.Extend()

	s.Empty(got)
}

func (s *DownCommandTestSuite) TestHandle() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down")

	app.EXPECT().StoragePath("framework/down").Return(tmpfile)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Info("The application is in maintenance mode now")

	cmd := NewDownCommand(app)
	err := cmd.Handle(mockContext)
	assert.Nil(s.T(), err)

	assert.True(s.T(), file.Exists(tmpfile))
}

func (s *DownCommandTestSuite) TestHandleWhenDownAlready() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down")

	_, err := os.Create(tmpfile)
	assert.Nil(s.T(), err)

	app.EXPECT().StoragePath("framework/down").Return(tmpfile)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Error("The application is in maintenance mode already!")

	cmd := NewDownCommand(app)
	err = cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}
