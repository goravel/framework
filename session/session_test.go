package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/str"
)

type SessionTestSuite struct {
	suite.Suite
	driver  *mocksession.Driver
	session *Session
	json    foundation.Json
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

func (s *SessionTestSuite) SetupTest() {
	s.driver = mocksession.NewDriver(s.T())
	s.json = json.NewJson()
	s.session = s.getSession()
}

func (s *SessionTestSuite) TestAll() {
	s.session.Put("key1", "value1").
		Put("key2", "value2")

	all := s.session.All()
	s.Equal("value1", all["key1"])
	s.Equal("value2", all["key2"])
}

func (s *SessionTestSuite) TestExists() {
	s.session.Put("foo", "bar")
	s.True(s.session.Exists("foo"))
	s.session.Put("baz", nil)
	s.False(s.session.Has("baz"))
	s.True(s.session.Exists("baz"))
	s.False(s.session.Exists("bogus"))
	s.True(s.session.Exists("foo"))
}

func (s *SessionTestSuite) TestFlash() {
	s.session.Flash("foo", "bar").
		Flash("bar", 0).
		Flash("baz", true)

	s.True(s.session.Has("foo"))
	s.Equal("bar", s.session.Get("foo"))
	s.Equal(0, s.session.Get("bar"))
	s.Equal(true, s.session.Get("baz"))

	s.session.ageFlashData()

	s.True(s.session.Has("foo"))
	s.Equal("bar", s.session.Get("foo"))
	s.Equal(0, s.session.Get("bar"))

	s.session.ageFlashData()

	s.False(s.session.Exists("foo"))
	s.Nil(s.session.Get("foo"))

	s.session.Flash("foo", "bar").
		Put("_flash.old", []any{"qu"})

	s.Equal([]any{"foo"}, s.session.Get("_flash.new"))
	s.Equal([]any{"qu"}, s.session.Get("_flash.old"))
}

func (s *SessionTestSuite) TestFlush() {
	s.session.Put("foo", "bar")
	s.session.Flush()
	s.False(s.session.Has("foo"))
}

func (s *SessionTestSuite) TestGet() {
	s.session.Put("key1", "value1")

	s.Equal("value1", s.session.Get("key1"))
	s.Nil(s.session.Get("key2"))
}

func (s *SessionTestSuite) TestGetID() {
	s.Equal(s.getSessionID(), s.session.GetID())
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

func (s *SessionTestSuite) TestInvalidate() {
	oldID := s.session.GetID()
	s.session.Put("foo", "bar")
	all := s.session.All()
	s.Equal("bar", all["foo"])

	s.session.Flash("name", "Krishan")
	s.True(s.session.Has("name"))

	s.driver.On("Destroy", oldID).Return(nil).Once()
	s.Nil(s.session.Invalidate())

	s.False(s.session.Has("name"))
	s.NotEqual(oldID, s.session.GetID())
	s.Equal(map[string]any{}, s.session.All())

	oldID = s.session.GetID()
	s.driver.On("Destroy", oldID).Return(errors.New("error")).Once()
	s.Equal(errors.New("error"), s.session.Invalidate())
	s.Equal(oldID, s.session.GetID())
}

func (s *SessionTestSuite) TestKeep() {
	s.session.Flash("name", "Krishan")
	s.session.Put("age", 22)
	s.session.Put("_flash.old", []any{"language"})
	s.Equal([]any{"name"}, s.session.Get("_flash.new"))

	s.session.Keep("name", "age", "language")
	s.Equal([]any{"name", "age", "language"}, s.session.Get("_flash.new"))
	s.Equal([]any{}, s.session.Get("_flash.old"))
}

func (s *SessionTestSuite) TestMigrate() {
	oldID := s.session.GetID()
	s.Nil(s.session.migrate())
	s.NotEqual(oldID, s.session.GetID())

	oldID = s.session.GetID()
	s.driver.On("Destroy", oldID).Return(nil).Once()
	s.Nil(s.session.migrate(true))
	s.NotEqual(oldID, s.session.GetID())

	// when driver is nil
	oldID = s.session.GetID()
	s.driver.On("Destroy", oldID).Return(nil).Once()
	s.session.SetDriver(nil)
	s.Nil(s.session.migrate(true))
	s.NotEqual(oldID, s.session.GetID())
}

func (s *SessionTestSuite) TestMissing() {
	s.session.Put("foo", "bar")
	s.False(s.session.Missing("foo"))
	s.session.Put("baz", nil)
	s.False(s.session.Has("baz"))
	s.False(s.session.Missing("baz"))
	s.True(s.session.Missing("bogus"))
}

func (s *SessionTestSuite) TestNow() {
	s.session.Now("foo", "bar")
	s.True(s.session.Has("foo"))
	s.Equal("bar", s.session.Get("foo"))

	s.session.ageFlashData()
	s.False(s.session.Has("foo"))
	s.Nil(s.session.Get("foo"))
}

func (s *SessionTestSuite) TestOnly() {
	s.session.
		Put("foo", "bar").
		Put("baz", "qux")

	all := s.session.All()
	s.Equal("bar", all["foo"])
	s.Equal("qux", all["baz"])
	s.Equal(map[string]any{"foo": "bar"}, s.session.Only([]string{"foo"}))
}

func (s *SessionTestSuite) TestPull() {
	s.session.Put("name", "Krishan")
	s.Equal("Krishan", s.session.Pull("name"))
	s.Equal("Krishan Kumar", s.session.Pull("name", "Krishan Kumar"))
	s.Nil(s.session.Get("name"))
}

func (s *SessionTestSuite) TestPut() {
	s.session.Put("key1", "value1")

	s.Equal("value1", s.session.Get("key1"))

	s.session.Put("key1", "value2")
	s.Equal("value2", s.session.Get("key1"))

	s.session.Put("key2", nil)
	s.Nil(s.session.Get("key2"))
}

func (s *SessionTestSuite) TestReflash() {
	s.session.Flash("foo", "bar").
		Put("_flash.old", []any{"foo"})

	s.session.Reflash()
	s.Equal([]any{"foo"}, s.session.Get("_flash.new"))
	s.Equal([]any{}, s.session.Get("_flash.old"))

	s.session.Now("foo", "bar")
	s.session.Reflash()
	s.Equal([]any{"foo"}, s.session.Get("_flash.new"))
	s.Equal([]any{}, s.session.Get("_flash.old"))
}

func (s *SessionTestSuite) TestRegenerate() {
	oldID := s.session.GetID()
	s.Nil(s.session.Regenerate())
	s.NotEqual(oldID, s.session.GetID())

	oldID = s.session.GetID()
	s.driver.On("Destroy", oldID).Return(nil).Once()
	s.Nil(s.session.Regenerate(true))
	s.NotEqual(oldID, s.session.GetID())

	oldID = s.session.GetID()
	s.driver.On("Destroy", oldID).Return(errors.New("error")).Once()
	s.Equal(errors.New("error"), s.session.Regenerate(true))
	s.Equal(oldID, s.session.GetID())
}

func (s *SessionTestSuite) TestRemove() {
	s.session.Put("foo", "bar")
	pulled := s.session.Remove("foo")
	s.Equal("bar", pulled)
	s.False(s.session.Has("foo"))
}

func (s *SessionTestSuite) TestSave() {
	s.driver.On("Read", s.getSessionID()).Return(``, nil).Once()
	s.session.Start()
	s.session.Put("key1", "value1").
		Flash("baz", "boom")

	data, _ := s.json.Marshal(map[string]any{
		"key1":       "value1",
		"baz":        "boom",
		"_token":     s.session.Token(),
		"_flash.new": []any{},
		"_flash.old": []any{"baz"},
	})
	s.driver.On("Write", s.getSessionID(), mock.MatchedBy(func(v string) bool {
		for _, key := range str.Of(string(data)).LTrim("{").RTrim("}").Split(",") {
			if !strings.Contains(v, key) {
				return false
			}
		}
		return true
	})).Return(nil).Once()

	s.Nil(s.session.Save())
	s.False(s.session.started)

	// there is an error when writing the json
	s.driver.On("Read", s.getSessionID()).Return(``, errors.New("error")).Once()
	s.session.Start()
	s.session.Put("key1", "value1")

	data, _ = s.json.Marshal(map[string]any{
		"key1":       "value1",
		"_token":     s.session.Token(),
		"_flash.new": []any{},
		"_flash.old": []any{},
	})
	s.driver.On("Write", s.getSessionID(), mock.MatchedBy(func(v string) bool {
		for _, key := range str.Of(string(data)).LTrim("{").RTrim("}").Split(",") {
			if !strings.Contains(v, key) {
				return false
			}
		}
		return true
	})).Return(errors.New("error")).Once()

	s.Equal(errors.New("error"), s.session.Save())
	s.True(s.session.started)
}

func (s *SessionTestSuite) TestSetID() {
	s.True(s.session.isValidID(s.session.GetID()))

	s.session.SetID("wrongId")
	s.NotEqual("wrongId", s.session.GetID())
	s.True(s.session.isValidID(s.session.GetID()))

	session := NewSession(s.getSessionName(), s.driver, s.json)
	s.True(session.isValidID(session.GetID()))
}

func (s *SessionTestSuite) TestSetName() {
	s.session.SetName("newName")
	s.Equal("newName", s.session.GetName())
}

func (s *SessionTestSuite) TestStart() {
	s.driver.On("Read", s.getSessionID()).Return(`{"foo":"bar"}`, nil).Once()
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
	s.driver.On("Read", s.getSessionID()).Return(`{"foo":"bar}`, nil).Once()
	s.session.Start()

	s.Nil(s.session.Get("foo"))
	s.Equal("qux", s.session.Get("baz"))
}

func (s *SessionTestSuite) TestToken() {
	s.driver.On("Read", s.getSessionID()).Return(`{"foo":"bar"}`, nil).Once()
	s.session.Start()

	s.True(len(s.session.Token()) == 40)
}

func (s *SessionTestSuite) TestRegenerateToken() {
	token := s.session.Get("_token")
	s.session.regenerateToken()
	s.NotEqual(token, s.session.Get("_token"))
}

func (s *SessionTestSuite) TestRemoveFromOldFlashData() {
	s.session.Put("foo", "bar").
		Put("baz", "qux").
		Put("_flash.old", []any{"foo", "baz"})

	s.session.removeFromOldFlashData("foo")
	s.Equal([]any{"baz"}, s.session.Get("_flash.old"))
}

func (s *SessionTestSuite) TestToStringSlice() {
	s.Equal([]string{"foo", "bar"}, toStringSlice([]any{"foo", "bar"}))

	s.Equal([]string{"1", "2", "3"}, toStringSlice([]any{"1", 2, "3"}))
}

func (s *SessionTestSuite) getSession() *Session {
	return NewSession(s.getSessionName(), s.driver, s.json, s.getSessionID())
}

func (s *SessionTestSuite) getSessionName() string {
	return "name"
}

func (s *SessionTestSuite) getSessionID() string {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
