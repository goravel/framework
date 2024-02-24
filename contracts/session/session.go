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
	// Put sets the value of a key in the session attributes.
	Put(key string, value any) Session
	// RegenerateToken regenerates the session token.
	RegenerateToken() Session
	// Forget removes specified keys from the session attributes.
	Forget(keys ...string) Session
}
