package session

// Session is the interface that defines the methods that should be implemented by a session.
type Session interface {
	// GetName returns the name of the session.
	GetName() string
	// SetName sets the name of the session.
	SetName(name string) Session
	// GetId returns the ID of the session.
	GetId() string
	// SetId sets the ID of the session.
	SetId(id string) Session
	// Start initiates the session.
	Start() bool
	// Save saves the session.
	Save() error
	// All returns all attributes of the session.
	All() map[string]any
	// Exists checks if a key exists in the session attributes.
	Exists(key string) bool
	// Missing checks if a key is missing in the session attributes.
	Missing(key string) bool
	// Has checks if a key exists and is not nil in the session attributes.
	Has(key string) bool
	// Get retrieves the value of a key from the session attributes.
	Get(key string, defaultValue ...any) any
	// Pull retrieves and removes the value of a key from the session attributes.
	Pull(key string, defaultValue ...any) any
	// Push adds a value to an array stored at the specified key in the session attributes.
	Push(key string, value any) Session
	// Put sets the value of a key in the session attributes.
	Put(key string, value any) Session
	// Token retrieves the session token.
	Token() string
	// RegenerateToken regenerates the session token.
	RegenerateToken() Session
	// Remove removes the value of a key from the session attributes.
	Remove(key string) any
	// Forget removes specified keys from the session attributes.
	Forget(keys ...string) Session
	// Flush clears all attributes from the session.
	Flush() Session
	// Flash sets a flash data value in the session attributes.
	Flash(key string, value any) Session
	// Invalidate invalidates the session.
	Invalidate() bool
	// Regenerate regenerates the session.
	Regenerate(destroy bool) bool
	// Only retrieves the specified keys and their values from the session attributes.
	Only(keys []string) map[string]any
	// Migrate migrates the session, optionally destroying the current session.
	Migrate(destroy bool) bool
	// PreviousUrl retrieves the previous URL stored in the session.
	PreviousUrl() string
	// SetPreviousUrl sets the previous URL in the session attributes.
	SetPreviousUrl(url string) Session
}
