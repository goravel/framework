package session

import (
	"maps"

	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/json"
	supportmaps "github.com/goravel/framework/support/maps"
	"github.com/goravel/framework/support/str"
)

type Session struct {
	id         string
	name       string
	attributes map[string]any
	driver     sessioncontract.Driver
	started    bool
}

func NewSession(name string, driver sessioncontract.Driver, id ...string) *Session {
	store := &Session{
		name:       name,
		driver:     driver,
		started:    false,
		attributes: make(map[string]any),
	}
	if len(id) > 0 {
		store.SetID(id[0])
	} else {
		store.SetID("")
	}

	return store
}

func (s *Session) All() map[string]any {
	return s.attributes
}

func (s *Session) Forget(keys ...string) sessioncontract.Session {
	supportmaps.Forget(s.attributes, keys...)

	return s
}

func (s *Session) Get(key string, defaultValue ...any) any {
	return supportmaps.Get(s.attributes, key, defaultValue...)
}

func (s *Session) GetID() string {
	return s.id
}

func (s *Session) GetName() string {
	return s.name
}

func (s *Session) Has(key string) bool {
	val, ok := s.attributes[key]
	if !ok {
		return false
	}

	return val != nil
}

func (s *Session) Put(key string, value any) sessioncontract.Session {
	s.attributes[key] = value
	return s
}

func (s *Session) RegenerateToken() sessioncontract.Session {
	return s.Put("_token", str.Random(40))
}

func (s *Session) Save() error {
	data, err := json.MarshalString(s.attributes)
	if err != nil {
		return err
	}

	if err = s.driver.Write(s.GetID(), data); err != nil {
		return err
	}

	s.started = false

	return nil
}

func (s *Session) SetID(id string) sessioncontract.Session {
	if s.isValidID(id) {
		s.id = id
	} else {
		s.id = s.generateSessionID()
	}

	return s
}

func (s *Session) SetName(name string) sessioncontract.Session {
	s.name = name

	return s
}

func (s *Session) Start() bool {
	s.loadSession()

	if !s.Has("_token") {
		s.RegenerateToken()
	}

	s.started = true
	return s.started
}

func (s *Session) generateSessionID() string {
	return str.Random(40)
}

func (s *Session) isValidID(id string) bool {
	return len(id) == 40
}

func (s *Session) loadSession() {
	data := s.readFromHandler()
	if data != nil {
		maps.Copy(s.attributes, data)
	}
}

func (s *Session) readFromHandler() map[string]any {
	value, err := s.driver.Read(s.GetID())
	if err != nil {
		return nil
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil
	}
	return data
}
