package session

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	mocksession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/json"
)

type StoreTestSuite struct {
	suite.Suite
	ctx     context.Context
	handler *mocksession.Handler
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{})
}

func (s *StoreTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.handler = mocksession.NewHandler(s.T())
}

func (s *StoreTestSuite) TestAll() {
	store := s.getSession()
	store.Put("key1", "value1").
		Put("key2", "value2")

	s.Equal(store.attributes, store.All())
}

func (s *StoreTestSuite) TestGet() {
	store := s.getSession()
	store.Put("key1", "value1")

	s.Equal("value1", store.Get("key1"))
	s.Nil(store.Get("key2"))
}

func (s *StoreTestSuite) TestGetId() {
	store := s.getSession()
	s.Equal(s.getSessionId(), store.GetId())
}

func (s *StoreTestSuite) TestGetName() {
	store := s.getSession()
	s.Equal(s.getSessionName(), store.GetName())
}

func (s *StoreTestSuite) TestForget() {
	store := s.getSession()
	store.Put("key1", "value1")
	store.Forget("key1")
	s.False(store.Has("key1"))

	store.Put("key2", "value2").
		Put("nilKey", nil)
	store.Forget("nilKey", "key2")
	s.False(store.Has("key2"))
	s.False(store.Has("nilKey"))
}

func (s *StoreTestSuite) TestHas() {
	store := s.getSession()
	store.Put("key1", "value1").
		Put("nilKey", nil)

	s.True(store.Has("key1"))
	s.False(store.Has("key2"))
	s.False(store.Has("nilKey"))
}

func (s *StoreTestSuite) TestPut() {
	store := s.getSession()
	store.Put("key1", "value1")

	s.Equal("value1", store.Get("key1"))

	store.Put("key1", "value2")
	s.Equal("value2", store.Get("key1"))

	store.Put("key2", nil)
	s.Nil(store.Get("key2"))
}

func (s *StoreTestSuite) TestRegenerateToken() {
	store := s.getSession()
	token := store.Get("_token")
	store.RegenerateToken()
	s.NotEqual(token, store.Get("_token"))
}

func (s *StoreTestSuite) TestSave() {
	store := s.getSession()
	s.handler.On("Read", s.getSessionId()).Once().Return(``)
	store.Start()
	store.Put("key1", "value1").
		Forget("_token")

	data, _ := json.MarshalString(store.attributes)
	s.handler.On("Write", s.getSessionId(), data).Once().Return(nil)

	s.Nil(store.Save())
	s.False(store.started)

	// there is an error when writing the json
	store = s.getSession()
	s.handler.On("Read", s.getSessionId()).Once().Return(``)
	store.Start()
	store.Put("key1", "value1").
		Forget("_token")

	data, _ = json.MarshalString(store.attributes)
	s.handler.On("Write", s.getSessionId(), data).Once().Return(errors.New("error"))

	s.Equal(errors.New("error"), store.Save())
	s.True(store.started)
}

func (s *StoreTestSuite) TestSetId() {
	store := s.getSession()
	s.True(store.isValidId(store.GetId()))

	store.SetId("wrongId")
	s.NotEqual("wrongId", store.GetId())
	s.True(store.isValidId(store.GetId()))

	store = NewStore(s.getSessionName(), s.handler)
	s.True(store.isValidId(store.GetId()))
}

func (s *StoreTestSuite) TestSetName() {
	store := s.getSession()
	store.SetName("newName")
	s.Equal("newName", store.GetName())
}

func (s *StoreTestSuite) TestStart() {
	store := s.getSession()
	s.handler.On("Read", s.getSessionId()).Once().Return(`{"foo":"bar"}`)
	store.Start()

	s.Equal("bar", store.Get("foo"))
	s.Equal("baz", store.Get("bar", "baz"))
	s.True(store.Has("foo"))
	s.False(store.Has("bar"))
	s.True(store.started)

	store.Put("baz", "qux")
	s.True(store.Has("baz"))

	// there is an error when parsing the json
	store = s.getSession()
	store.Put("baz", "qux")
	s.handler.On("Read", s.getSessionId()).Once().Return(`{"foo":"bar}`)
	store.Start()

	s.Nil(store.Get("foo"))
	s.Equal("qux", store.Get("baz"))
}

func (s *StoreTestSuite) getSession() *Store {
	return NewStore(s.getSessionName(), s.handler, s.getSessionId())
}

func (s *StoreTestSuite) getSessionName() string {
	return "name"
}

func (s *StoreTestSuite) getSessionId() string {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
