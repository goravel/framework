package session

import (
	"maps"
	"slices"

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

func (s *Session) Exists(key string) bool {
	return supportmaps.Exists(s.attributes, key)
}

func (s *Session) Flash(key string, value any) sessioncontract.Session {
	s.Put(key, value)

	old := s.Get("_flash.new", []string{}).([]string)
	s.Put("_flash.new", append(old, key))

	s.removeFromOldFlashData(key)

	return s
}

func (s *Session) Flush() sessioncontract.Session {
	s.attributes = make(map[string]any)
	return s
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

func (s *Session) Invalidate() error {
	s.Flush()
	return s.Migrate(true)
}

func (s *Session) Migrate(destroy ...bool) error {
	shouldDestroy := false
	if len(destroy) > 0 {
		shouldDestroy = destroy[0]
	}

	if shouldDestroy {
		err := s.driver.Destroy(s.GetID())
		if err != nil {
			return err
		}
	}

	s.SetID(s.generateSessionID())

	return nil
}

func (s *Session) Missing(key string) bool {
	return !s.Exists(key)
}

func (s *Session) Only(keys []string) map[string]any {
	return supportmaps.Only(s.attributes, keys...)
}

func (s *Session) PreviousUrl() string {
	return s.Get("_previous.url").(string)
}

func (s *Session) Pull(key string, def ...any) any {
	return supportmaps.Pull(s.attributes, key, def...)
}

func (s *Session) Put(key string, value any) sessioncontract.Session {
	s.attributes[key] = value
	return s
}

func (s *Session) Regenerate(destroy ...bool) error {
	err := s.Migrate(destroy...)
	if err != nil {
		return err
	}

	s.RegenerateToken()
	return nil
}

func (s *Session) RegenerateToken() sessioncontract.Session {
	return s.Put("_token", str.Random(40))
}

func (s *Session) Remove(key string) any {
	return s.Pull(key)
}

func (s *Session) Save() error {
	s.ageFlashData()

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

func (s *Session) SetPreviousUrl(url string) sessioncontract.Session {
	return s.Put("_previous.url", url)
}

func (s *Session) Start() bool {
	s.loadSession()

	if !s.Has("_token") {
		s.RegenerateToken()
	}

	s.started = true
	return s.started
}

func (s *Session) Token() string {
	return s.Get("_token").(string)
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

func (s *Session) ageFlashData() {
	old := s.Get("_flash.old", []string{}).([]string)

	s.Forget(old...)
	s.Put("_flash.old", s.Get("_flash.new", []string{}))
	s.Put("_flash.new", []string{})
}

func (s *Session) removeFromOldFlashData(keys ...string) {
	old := s.Get("_flash.old", []string{}).([]string)
	for _, key := range keys {
		old = slices.DeleteFunc(old, func(i string) bool {
			return i == key
		})
	}
	s.Put("_flash.old", old)
}
