package session

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/json"
)

type SessionTestSuite struct {
	suite.Suite
	driver  *mocksession.Driver
	session *Session
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

func (s *SessionTestSuite) SetupTest() {
	s.driver = mocksession.NewDriver(s.T())
	s.session = s.getSession()
}

func (s *SessionTestSuite) TestAll() {
	s.session.Put("key1", "value1").
		Put("key2", "value2")

	all := s.session.All()
	s.Equal("value1", all["key1"])
	s.Equal("value2", all["key2"])
}

func (s *SessionTestSuite) TestGet() {
	s.session.Put("key1", "value1")

	s.Equal("value1", s.session.Get("key1"))
	s.Nil(s.session.Get("key2"))
}

func (s *SessionTestSuite) TestGetID() {
	s.Equal(s.getSessionId(), s.session.GetID())
}

func (s *SessionTestSuite) TestGetName() {
	s.Equal(s.getSessionName(), s.session.GetName())
}

func (s *SessionTestSuite) TestForget() {
	s.session.Put("key1", "value1")
	s.session.Forget("key1")
	s.False(s.session.Has("key1"))

	s.session.Put("key2", "value2").
		Put("nilKey", nil)
	s.session.Forget("nilKey", "key2")
	s.False(s.session.Has("key2"))
	s.False(s.session.Has("nilKey"))
}

func (s *SessionTestSuite) TestHas() {
	s.session.Put("key1", "value1").
		Put("nilKey", nil)

	s.True(s.session.Has("key1"))
	s.False(s.session.Has("key2"))
	s.False(s.session.Has("nilKey"))
}

func (s *SessionTestSuite) TestPut() {
	s.session.Put("key1", "value1")

	s.Equal("value1", s.session.Get("key1"))

	s.session.Put("key1", "value2")
	s.Equal("value2", s.session.Get("key1"))

	s.session.Put("key2", nil)
	s.Nil(s.session.Get("key2"))
}

func (s *SessionTestSuite) TestRegenerateToken() {
	token := s.session.Get("_token")
	s.session.RegenerateToken()
	s.NotEqual(token, s.session.Get("_token"))
}

func (s *SessionTestSuite) TestSave() {
	s.driver.On("Read", s.getSessionId()).Return(``, nil).Once()
	s.session.Start()
	s.session.Put("key1", "value1").
		Forget("_token")

	data, _ := json.MarshalString(s.session.All())
	s.driver.On("Write", s.getSessionId(), data).Return(nil).Once()

	s.Nil(s.session.Save())
	s.False(s.session.started)

	// there is an error when writing the json
	s.driver.On("Read", s.getSessionId()).Return(``, errors.New("error")).Once()
	s.session.Start()
	s.session.Put("key1", "value1").
		Forget("_token")

	data, _ = json.MarshalString(s.session.All())
	s.driver.On("Write", s.getSessionId(), data).Return(errors.New("error")).Once()

	s.Equal(errors.New("error"), s.session.Save())
	s.True(s.session.started)
}

func (s *SessionTestSuite) TestSetID() {
	s.True(s.session.isValidID(s.session.GetID()))

	s.session.SetID("wrongId")
	s.NotEqual("wrongId", s.session.GetID())
	s.True(s.session.isValidID(s.session.GetID()))

	session := NewSession(s.getSessionName(), s.driver)
	s.True(session.isValidID(session.GetID()))
}

func (s *SessionTestSuite) TestSetName() {
	s.session.SetName("newName")
	s.Equal("newName", s.session.GetName())
}

func (s *SessionTestSuite) TestStart() {
	s.driver.On("Read", s.getSessionId()).Return(`{"foo":"bar"}`, nil).Once()
	s.session.Start()

	s.Equal("bar", s.session.Get("foo"))
	s.Equal("baz", s.session.Get("bar", "baz"))
	s.True(s.session.Has("foo"))
	s.False(s.session.Has("bar"))
	s.True(s.session.started)

	s.session.Put("baz", "qux")
	s.True(s.session.Has("baz"))

	// there is an error when parsing the json
	s.session = s.getSession()
	s.session.Put("baz", "qux")
	s.driver.On("Read", s.getSessionId()).Return(`{"foo":"bar}`, nil).Once()
	s.session.Start()

	s.Nil(s.session.Get("foo"))
	s.Equal("qux", s.session.Get("baz"))
}

func (s *SessionTestSuite) getSession() *Session {
	return NewSession(s.getSessionName(), s.driver, s.getSessionId())
}

func (s *SessionTestSuite) getSessionName() string {
	return "name"
}

func (s *SessionTestSuite) getSessionId() string {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
