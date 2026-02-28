package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshash "github.com/goravel/framework/mocks/hash"
	mocksview "github.com/goravel/framework/mocks/view"
)

type DownCommandTestSuite struct {
	suite.Suite
	mockApp     *mocksfoundation.Application
	mockCtx     *mocksconsole.Context
	mockHash    *mockshash.Hash
	mockView    *mocksview.View
	mockStorage *mocksfilesystem.Storage
}

func TestDownCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DownCommandTestSuite))
}

func (s *DownCommandTestSuite) SetupSuite() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.mockCtx = mocksconsole.NewContext(s.T())
	s.mockHash = mockshash.NewHash(s.T())
	s.mockView = mocksview.NewView(s.T())
	s.mockStorage = mocksfilesystem.NewStorage(s.T())

	s.mockApp.EXPECT().MakeHash().Return(s.mockHash)
	s.mockApp.EXPECT().MakeView().Return(s.mockView)
	s.mockApp.EXPECT().MakeStorage().Return(s.mockStorage)
}

func (s *DownCommandTestSuite) TestSignature() {
	expected := "down"
	s.Require().Equal(expected, NewDownCommand(s.mockApp).Signature())
}

func (s *DownCommandTestSuite) TestDescription() {
	expected := "Put the application into maintenance mode"
	s.Require().Equal(expected, NewDownCommand(s.mockApp).Description())
}

func (s *DownCommandTestSuite) TestExtend() {
	cmd := NewDownCommand(s.mockApp)
	got := cmd.Extend()

	s.Equal(6, len(got.Flags))
}

func (s *DownCommandTestSuite) TestHandle() {
	cmd := NewDownCommand(s.mockApp)

	flag := cmd.Extend().Flags[0].(*command.StringFlag)
	s.mockCtx.EXPECT().OptionInt("status").Return(503)
	s.mockCtx.EXPECT().Option("render").Return("")
	s.mockCtx.EXPECT().Option("redirect").Return("")
	s.mockCtx.EXPECT().Option("secret").Return("")
	s.mockCtx.EXPECT().OptionBool("with-secret").Return(false)
	s.mockCtx.EXPECT().Option("reason").Return(flag.Value)
	s.mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"The application is under maintenance\",\"status\":503}").Return(nil)

	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithReason() {
	s.mockCtx.EXPECT().OptionInt("status").Return(505)
	// s.mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	// s.mockCtx.EXPECT().Option("redirect").Return("").Once()
	// s.mockCtx.EXPECT().Option("render").Return("").Once()
	// s.mockCtx.EXPECT().Option("secret").Return("").Once()
	// s.mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	// s.mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":505}").Return(nil)

	cmd := NewDownCommand(s.mockApp)
	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithRedirect() {
	s.mockCtx.EXPECT().OptionInt("status").Return(503)
	s.mockCtx.EXPECT().Option("render").Return("")
	s.mockCtx.EXPECT().Option("redirect").Return("/maintenance")
	s.mockCtx.EXPECT().Option("secret").Return("")
	s.mockCtx.EXPECT().OptionBool("with-secret").Return(false)
	s.mockCtx.EXPECT().Success("The application is in maintenance mode now")
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":503, \"redirect\": \"/maintenance\"}").Return(nil)

	cmd := NewDownCommand(s.mockApp)
	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithRender() {
	s.mockView.EXPECT().Exists("errors/503.tmpl").Return(true)

	s.mockCtx.EXPECT().OptionInt("status").Return(503)
	s.mockCtx.EXPECT().Option("render").Return("errors/503.tmpl")
	s.mockCtx.EXPECT().Option("redirect").Return("")
	s.mockCtx.EXPECT().Option("secret").Return("")
	s.mockCtx.EXPECT().OptionBool("with-secret").Return(false)
	s.mockCtx.EXPECT().Success("The application is in maintenance mode now")
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":503, \"render\": \"errors/503.tmpl\"}").Return(nil)

	cmd := NewDownCommand(s.mockApp)
	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleSecret() {
	s.mockHash.EXPECT().Make("secretpassword").Return("hashedsecretpassword", nil)

	s.mockCtx.EXPECT().OptionInt("status").Return(503)
	s.mockCtx.EXPECT().Option("reason").Return("Under maintenance")
	s.mockCtx.EXPECT().Option("render").Return("")
	s.mockCtx.EXPECT().Option("redirect").Return("")
	s.mockCtx.EXPECT().Option("secret").Return("secretpassword").Once()
	s.mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	s.mockCtx.EXPECT().Success("The application is in maintenance mode now")
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":503, \"secret\": \"hashedsecretpassword\"}").Return(nil)

	cmd := NewDownCommand(s.mockApp)
	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithSecret() {
	s.mockHash.EXPECT().Make(mock.Anything).Return("randomhashedsecretpassword", nil)

	s.mockCtx.EXPECT().OptionInt("status").Return(503)
	s.mockCtx.EXPECT().Option("reason").Return("Under maintenance")
	s.mockCtx.EXPECT().Option("render").Return("")
	s.mockCtx.EXPECT().Option("redirect").Return("")
	s.mockCtx.EXPECT().Option("secret").Return("")
	s.mockCtx.EXPECT().OptionBool("with-secret").Return(true)
	s.mockCtx.EXPECT().Info("Using secret: randomhashedsecretpassword").Once()
	s.mockCtx.EXPECT().Success("The application is in maintenance mode now")
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":503, \"secret\": \"randomhashedsecretpassword\"}").Return(nil).Once()

	cmd := NewDownCommand(s.mockApp)
	err := cmd.Handle(s.mockCtx)

	assert.Nil(s.T(), err)
}
