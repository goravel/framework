package console

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
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

	s.Equal(6, len(got.Flags))
}

func (s *DownCommandTestSuite) TestHandle() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/maintenance")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	cmd := NewDownCommand(app)

	mockContext := mocksconsole.NewContext(s.T())
	flag := cmd.Extend().Flags[0].(*command.StringFlag)
	mockContext.EXPECT().OptionInt("status").Return(503)
	mockContext.EXPECT().Option("render").Return("")
	mockContext.EXPECT().Option("redirect").Return("")
	mockContext.EXPECT().Option("secret").Return("")
	mockContext.EXPECT().OptionBool("with-secret").Return(false)
	mockContext.EXPECT().Option("reason").Return(flag.Value)
	mockContext.EXPECT().Info("The application is in maintenance mode now")

	err := cmd.Handle(mockContext)

	assert.Nil(s.T(), err)
	assert.True(s.T(), file.Exists(tmpfile))

	content, err := file.GetContent(tmpfile)

	assert.Nil(s.T(), err)

	var maintenanceOptions *MaintenanceOptions
	err = json.Unmarshal([]byte(content), &maintenanceOptions)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), flag.Value, maintenanceOptions.Reason)
	assert.Equal(s.T(), 503, maintenanceOptions.Status)
}

func (s *DownCommandTestSuite) TestHandleWithReason() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down_with_reason")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().OptionInt("status").Return(505)
	mockContext.EXPECT().Option("render").Return("")
	mockContext.EXPECT().Option("redirect").Return("")
	mockContext.EXPECT().Option("secret").Return("")
	mockContext.EXPECT().OptionBool("with-secret").Return(false)
	mockContext.EXPECT().Option("reason").Return("Under maintenance")
	mockContext.EXPECT().Info("The application is in maintenance mode now")

	cmd := NewDownCommand(app)
	err := cmd.Handle(mockContext)

	assert.Nil(s.T(), err)
	assert.True(s.T(), file.Exists(tmpfile))

	content, err := file.GetContent(tmpfile)

	assert.Nil(s.T(), err)
	var maintenanceOptions *MaintenanceOptions
	err = json.Unmarshal([]byte(content), &maintenanceOptions)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "Under maintenance", maintenanceOptions.Reason)
	assert.Equal(s.T(), 505, maintenanceOptions.Status)
}
