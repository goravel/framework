package session

type Session interface {
	GetName() string
	SetName(name string) Session
	GetId() string
	SetId(id string) Session
	Start() bool
	Save() error
	All() map[string]any
	Exists(key string) bool
	Missing(key string) bool
	Has(key string) bool
	Get(key string, defaultValue ...any) any
	Pull(key string, defaultValue ...any) any
	Push(key string, value any) Session
	Put(key string, value any) Session
	Token() string
	RegenerateToken() Session
	Remove(key string) any
	Forget(keys ...string) Session
	Flush() Session
	Flash(key string, value any) Session
	Invalidate() bool
	Regenerate(destroy bool) bool
	Only(keys []string) map[string]any
	Migrate(destroy bool) bool
	PreviousUrl() string
	SetPreviousUrl(url string) Session
}
