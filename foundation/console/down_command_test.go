package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
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
	app.EXPECT().StoragePath("framework/down").Return("/tmp/down")

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Info("The application is in maintenance mode now")

	cmd := NewDownCommand(app)
	cmd.Handle(mockContext)
	os.Remove("/tmp/down")
}

func (s *DownCommandTestSuite) TestHandleWhenDownAlready() {
	app := mocksfoundation.NewApplication(s.T())
	app.EXPECT().StoragePath("framework/down").Return("/tmp/down")
	os.Create("/tmp/down")

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Error("The application is in maintenance mode already!")

	cmd := NewDownCommand(app)
	cmd.Handle(mockContext)
	os.Remove("/tmp/down")
}
