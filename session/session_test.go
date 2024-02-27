package session

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/json"
)

type SessionTestSuite struct {
	suite.Suite
	ctx    context.Context
	driver *mocksession.Driver
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

func (s *SessionTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.driver = mocksession.NewDriver(s.T())
}

func (s *SessionTestSuite) TestAll() {
	session := s.getSession()
	session.Put("key1", "value1").
		Put("key2", "value2")

	s.Equal(session.attributes, session.All())
}

func (s *SessionTestSuite) TestGet() {
	session := s.getSession()
	session.Put("key1", "value1")

	s.Equal("value1", session.Get("key1"))
	s.Nil(session.Get("key2"))
}

func (s *SessionTestSuite) TestGetID() {
	session := s.getSession()
	s.Equal(s.getSessionId(), session.GetID())
}

func (s *SessionTestSuite) TestGetName() {
	session := s.getSession()
	s.Equal(s.getSessionName(), session.GetName())
}

func (s *SessionTestSuite) TestForget() {
	session := s.getSession()
	session.Put("key1", "value1")
	session.Forget("key1")
	s.False(session.Has("key1"))

	session.Put("key2", "value2").
		Put("nilKey", nil)
	session.Forget("nilKey", "key2")
	s.False(session.Has("key2"))
	s.False(session.Has("nilKey"))
}

func (s *SessionTestSuite) TestHas() {
	session := s.getSession()
	session.Put("key1", "value1").
		Put("nilKey", nil)

	s.True(session.Has("key1"))
	s.False(session.Has("key2"))
	s.False(session.Has("nilKey"))
}

func (s *SessionTestSuite) TestPut() {
	session := s.getSession()
	session.Put("key1", "value1")

	s.Equal("value1", session.Get("key1"))

	session.Put("key1", "value2")
	s.Equal("value2", session.Get("key1"))

	session.Put("key2", nil)
	s.Nil(session.Get("key2"))
}

func (s *SessionTestSuite) TestRegenerateToken() {
	session := s.getSession()
	token := session.Get("_token")
	session.RegenerateToken()
	s.NotEqual(token, session.Get("_token"))
}

func (s *SessionTestSuite) TestSave() {
	session := s.getSession()
	s.driver.On("Read", s.getSessionId()).Once().Return(``)
	session.Start()
	session.Put("key1", "value1").
		Forget("_token")

	data, _ := json.MarshalString(session.attributes)
	s.driver.On("Write", s.getSessionId(), data).Once().Return(nil)

	s.Nil(session.Save())
	s.False(session.started)

	// there is an error when writing the json
	session = s.getSession()
	s.driver.On("Read", s.getSessionId()).Once().Return(``)
	session.Start()
	session.Put("key1", "value1").
		Forget("_token")

	data, _ = json.MarshalString(session.attributes)
	s.driver.On("Write", s.getSessionId(), data).Once().Return(errors.New("error"))

	s.Equal(errors.New("error"), session.Save())
	s.True(session.started)
}

func (s *SessionTestSuite) TestSetID() {
	session := s.getSession()
	s.True(session.isValidID(session.GetID()))

	session.SetID("wrongId")
	s.NotEqual("wrongId", session.GetID())
	s.True(session.isValidID(session.GetID()))

	session = NewSession(s.getSessionName(), s.driver)
	s.True(session.isValidID(session.GetID()))
}

func (s *SessionTestSuite) TestSetName() {
	session := s.getSession()
	session.SetName("newName")
	s.Equal("newName", session.GetName())
}

func (s *SessionTestSuite) TestStart() {
	session := s.getSession()
	s.driver.On("Read", s.getSessionId()).Once().Return(`{"foo":"bar"}`)
	session.Start()

	s.Equal("bar", session.Get("foo"))
	s.Equal("baz", session.Get("bar", "baz"))
	s.True(session.Has("foo"))
	s.False(session.Has("bar"))
	s.True(session.started)

	session.Put("baz", "qux")
	s.True(session.Has("baz"))

	// there is an error when parsing the json
	session = s.getSession()
	session.Put("baz", "qux")
	s.driver.On("Read", s.getSessionId()).Once().Return(`{"foo":"bar}`)
	session.Start()

	s.Nil(session.Get("foo"))
	s.Equal("qux", session.Get("baz"))
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
