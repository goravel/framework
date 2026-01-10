package console

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshash "github.com/goravel/framework/mocks/hash"
	mockshttp "github.com/goravel/framework/mocks/http"
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

func (s *DownCommandTestSuite) TestHandleWithRedirect() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down_with_reason")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().OptionInt("status").Return(503)
	mockContext.EXPECT().Option("render").Return("")
	mockContext.EXPECT().Option("redirect").Return("/maintenance")
	mockContext.EXPECT().Option("secret").Return("")
	mockContext.EXPECT().OptionBool("with-secret").Return(false)
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

	assert.Equal(s.T(), "/maintenance", maintenanceOptions.Redirect)
	assert.Equal(s.T(), 503, maintenanceOptions.Status)
}

func (s *DownCommandTestSuite) TestHandleWithRender() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down_with_reason")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	views := NewView(s.T())
	views.EXPECT().Exists("errors/503.tmpl").Return(true)
	app.EXPECT().MakeView().Return(views)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().OptionInt("status").Return(503)
	mockContext.EXPECT().Option("render").Return("errors/503.tmpl")
	mockContext.EXPECT().Option("redirect").Return("")
	mockContext.EXPECT().Option("secret").Return("")
	mockContext.EXPECT().OptionBool("with-secret").Return(false)
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

	assert.Equal(s.T(), "errors/503.tmpl", maintenanceOptions.Render)
	assert.Equal(s.T(), 503, maintenanceOptions.Status)
}

func (s *DownCommandTestSuite) TestHandleSecret() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down_with_reason")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	hash := mockshash.NewHash(s.T())
	hash.EXPECT().Make("secretpassword").Return("hashedsecretpassword", nil)
	app.EXPECT().MakeHash().Return(hash)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().OptionInt("status").Return(503)
	mockContext.EXPECT().Option("reason").Return("Under maintenance")
	mockContext.EXPECT().Option("render").Return("")
	mockContext.EXPECT().Option("redirect").Return("")
	mockContext.EXPECT().Option("secret").Return("secretpassword")
	mockContext.EXPECT().OptionBool("with-secret").Return(false)
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
	assert.Equal(s.T(), "hashedsecretpassword", maintenanceOptions.Secret)
	assert.Equal(s.T(), 503, maintenanceOptions.Status)
}

func (s *DownCommandTestSuite) TestHandleWithSecret() {
	app := mocksfoundation.NewApplication(s.T())
	tmpfile := filepath.Join(s.T().TempDir(), "/down_with_reason")

	app.EXPECT().StoragePath("framework/maintenance").Return(tmpfile)

	hash := mockshash.NewHash(s.T())
	hash.EXPECT().Make(mock.Anything).Return("randomhashedsecretpassword", nil)
	app.EXPECT().MakeHash().Return(hash)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().OptionInt("status").Return(503)
	mockContext.EXPECT().Option("reason").Return("Under maintenance")
	mockContext.EXPECT().Option("render").Return("")
	mockContext.EXPECT().Option("redirect").Return("")
	mockContext.EXPECT().Option("secret").Return("")
	mockContext.EXPECT().OptionBool("with-secret").Return(true)
	mockContext.EXPECT().Info(mock.MatchedBy(func(msg string) bool {
		return strings.HasPrefix(msg, "Using secret: ")
	}))
	mockContext.EXPECT().Info("The application is in maintenance mode now")

	cmd := NewDownCommand(app)
	err := cmd.Handle(mockContext)

	assert.Nil(s.T(), err)
	assert.True(s.T(), file.Exists(tmpfile))

	content, err := os.ReadFile(tmpfile)

	assert.Nil(s.T(), err)
	var maintenanceOptions *MaintenanceOptions
	err = json.Unmarshal(content, &maintenanceOptions)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "Under maintenance", maintenanceOptions.Reason)
	assert.Equal(s.T(), "randomhashedsecretpassword", maintenanceOptions.Secret)
	assert.Equal(s.T(), 503, maintenanceOptions.Status)
}

func NewView(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockshttp.View {
	mock := &mockshttp.View{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
