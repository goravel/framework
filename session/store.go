package session

import (
	"maps"
	"slices"

	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/json"
	"github.com/goravel/framework/support/str"
)

type Store struct {
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
	s.ageFlashData()

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

func (s *Store) Exists(key string) bool {
	_, ok := s.attributes[key]
	return ok
}

func (s *Store) Missing(key string) bool {
	return !s.Exists(key)
}

func (s *Store) Has(key string) bool {
	val, ok := s.attributes[key]
	if !ok {
		return false
	}

	return val != nil
}

func (s *Store) Get(key string, defaultValue ...any) any {
	val, ok := s.attributes[key]
	if !ok && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return val
}

func (s *Store) Pull(key string, def ...any) any {
	if val, ok := s.attributes[key]; ok {
		delete(s.attributes, key)
		return val
	}

	if len(def) > 0 {
		return def[0]
	}

	return nil
}

func (s *Store) Push(key string, value any) sessioncontract.Session {
	arr := s.Get(key, make([]any, 0)).([]any)
	arr = append(arr, value)
	return s.Put(key, arr)
}

func (s *Store) Put(key string, value any) sessioncontract.Session {
	s.attributes[key] = value
	return s
}

func (s *Store) Token() string {
	return s.Get("_token").(string)
}

func (s *Store) RegenerateToken() sessioncontract.Session {
	return s.Put("_token", str.Random(40))
}

func (s *Store) Remove(key string) any {
	return s.Pull(key)
}

func (s *Store) Forget(keys ...string) sessioncontract.Session {
	for _, key := range keys {
		delete(s.attributes, key)
	}

	return s
}

func (s *Store) Flush() sessioncontract.Session {
	s.attributes = make(map[string]any)
	return s
}

func (s *Store) Flash(key string, value any) sessioncontract.Session {
	s.Put(key, value)
	s.Push("_flash.new", key)
	s.removeFromOldFlashData(key)

	return s
}

func (s *Store) Only(keys []string) map[string]any {
	result := make(map[string]any)
	for _, key := range keys {
		if val, ok := s.attributes[key]; ok {
			result[key] = val
		}
	}

	return result
}

func (s *Store) Invalidate() bool {
	s.Flush()
	return s.Migrate(true)
}

func (s *Store) Regenerate(destroy bool) bool {
	return true
}

func (s *Store) Migrate(destroy bool) bool {
	if destroy {
		s.handler.Destroy(s.GetId())
	}

	s.SetId(s.generateSessionId())

	return true
}

func (s *Store) PreviousUrl() string {
	return s.Get("_previous.url").(string)
}

func (s *Store) SetPreviousUrl(url string) sessioncontract.Session {
	return s.Put("_previous.url", url)
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

func (s *Store) ageFlashData() {
	s.Forget(s.Get("_flash.old", make([]string, 0)).([]string)...)
	s.Put("_flash.old", s.Get("_flash.new", make([]any, 0)))
	s.Put("_flash.new", make([]any, 0))
}

func (s *Store) removeFromOldFlashData(keys ...string) {
	old := s.Get("_flash.old", make([]any, 0)).([]any)
	for _, key := range keys {
		old = slices.DeleteFunc(old, func(i any) bool {
			return i == key
		})
	}
	s.Put("_flash.old", old)
}
