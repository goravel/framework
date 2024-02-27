package session

// Session is the interface that defines the methods that should be implemented by a session.
type Session interface {
	// All returns all attributes of the session.
	All() map[string]any
	// Forget removes specified keys from the session attributes.
	Forget(keys ...string) Session
	// Get retrieves the value of a key from the session attributes.
	Get(key string, defaultValue ...any) any
	// GetName returns the name of the session.
	GetName() string
	// GetID returns the ID of the session.
	GetID() string
	// Has checks if a key exists and is not nil in the session attributes.
	Has(key string) bool
	// Put sets the value of a key in the session attributes.
	Put(key string, value any) Session
	// RegenerateToken regenerates the session token.
	RegenerateToken() Session
	// Save saves the session.
	Save() error
	// SetID sets the ID of the session.
	SetID(id string) Session
	// SetName sets the name of the session.
	SetName(name string) Session
	// Start initiates the session.
	Start() bool
}
