package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type UpCommandTestSuite struct {
	suite.Suite
}

func TestUpCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UpCommandTestSuite))
}

func (s *UpCommandTestSuite) SetupSuite() {
}

func (s *UpCommandTestSuite) TearDownSuite() {
}

func (s *UpCommandTestSuite) TestSignature() {
	expected := "up"
	s.Require().Equal(expected, NewUpCommand(mocksfoundation.NewApplication(s.T())).Signature())
}

func (s *UpCommandTestSuite) TestDescription() {
	expected := "Bring the application out of maintenance mode"
	s.Require().Equal(expected, NewUpCommand(mocksfoundation.NewApplication(s.T())).Description())
}

func (s *UpCommandTestSuite) TestExtend() {
	cmd := NewUpCommand(mocksfoundation.NewApplication(s.T()))
	got := cmd.Extend()

	s.Empty(got)
}

func (s *UpCommandTestSuite) TestHandle() {
	app := mocksfoundation.NewApplication(s.T())
	file, err := os.CreateTemp("", "down")
	println(file.Name(), err)

	app.EXPECT().StoragePath("framework/down").Return(file.Name())

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Info("The application is up and live now")

	cmd := NewUpCommand(app)
	err = cmd.Handle(mockContext)
	assert.Nil(s.T(), err)
}

func (s *UpCommandTestSuite) TestHandleWhenNotDown() {
	app := mocksfoundation.NewApplication(s.T())
	app.EXPECT().StoragePath("framework/down").Return(os.TempDir() + "/down")

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Error("The application is not in maintenance mode")

	cmd := NewUpCommand(app)
	cmd.Handle(mockContext)
}
