package session

import (
	"context"
	"maps"

	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/json"
	supportmaps "github.com/goravel/framework/support/maps"
	"github.com/goravel/framework/support/str"
)

type Store struct {
	ctx        context.Context
	id         string
	name       string
	attributes map[string]any
	handler    sessioncontract.Handler
	started    bool
}

func NewStore(name string, handler sessioncontract.Handler, id ...string) *Store {
	store := &Store{
		name:       name,
		handler:    handler,
		started:    false,
		attributes: make(map[string]any),
	}
	if len(id) > 0 {
		store.SetId(id[0])
	} else {
		store.SetId("")
	}

	return store
}

func (s *Store) GetName() string {
	return s.name
}

func (s *Store) SetName(name string) sessioncontract.Session {
	s.name = name

	return s
}

func (s *Store) GetId() string {
	return s.id
}

func (s *Store) SetId(id string) sessioncontract.Session {
	if s.isValidId(id) {
		s.id = id
	} else {
		s.id = s.generateSessionId()
	}

	return s
}

func (s *Store) Start() bool {
	s.loadSession()

	if !s.Has("_token") {
		s.RegenerateToken()
	}

	s.started = true
	return s.started
}

func (s *Store) Save() error {
	data, err := json.MarshalString(s.attributes)
	if err != nil {
		return err
	}

	if err = s.handler.Write(s.GetId(), data); err != nil {
		return err
	}

	s.started = false

	return nil
}

func (s *Store) All() map[string]any {
	return s.attributes
}

func (s *Store) Has(key string) bool {
	val, ok := s.attributes[key]
	if !ok {
		return false
	}

	return val != nil
}

func (s *Store) Get(key string, defaultValue ...any) any {
	return supportmaps.Get(s.attributes, key, defaultValue...)
}

func (s *Store) Put(key string, value any) sessioncontract.Session {
	s.attributes[key] = value
	return s
}

func (s *Store) RegenerateToken() sessioncontract.Session {
	return s.Put("_token", str.Random(40))
}

func (s *Store) Forget(keys ...string) sessioncontract.Session {
	supportmaps.Forget(s.attributes, keys...)

	return s
}

func (s *Store) generateSessionId() string {
	return str.Random(40)
}

func (s *Store) isValidId(id string) bool {
	return len(id) == 40
}

func (s *Store) loadSession() {
	data := s.readFromHandler()
	if data != nil {
		maps.Copy(s.attributes, data)
	}
}

func (s *Store) readFromHandler() map[string]any {
	var data map[string]any
	if err := json.Unmarshal([]byte(s.handler.Read(s.GetId())), &data); err != nil {
		return nil
	}
	return data
}
