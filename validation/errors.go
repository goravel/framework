package validation

// Errors collects validation error messages.
// Structure: field -> rule -> message
type Errors struct {
	messages map[string]map[string]string
}

// NewErrors creates a new empty Errors.
func NewErrors() *Errors {
	return &Errors{
		messages: make(map[string]map[string]string),
	}
}

// Add adds an error message for a field and rule.
func (m *Errors) Add(field, rule, message string) {
	if _, ok := m.messages[field]; !ok {
		m.messages[field] = make(map[string]string)
	}
	m.messages[field][rule] = message
}

// One gets the first error message. If a key is provided, returns the first
// error for that field. Otherwise returns the first error overall.
func (m *Errors) One(key ...string) string {
	if len(key) > 0 && key[0] != "" {
		if fieldErrors, ok := m.messages[key[0]]; ok {
			for _, msg := range fieldErrors {
				return msg
			}
		}
		return ""
	}

	for _, fieldErrors := range m.messages {
		for _, msg := range fieldErrors {
			return msg
		}
	}
	return ""
}

// Get gets all error messages for a given field (rule -> message).
func (m *Errors) Get(key string) map[string]string {
	if errors, ok := m.messages[key]; ok {
		return errors
	}
	return nil
}

// All gets all error messages (field -> rule -> message).
func (m *Errors) All() map[string]map[string]string {
	return m.messages
}

// Has checks if there are any error messages for a given field.
func (m *Errors) Has(key string) bool {
	errors, ok := m.messages[key]
	return ok && len(errors) > 0
}

// IsEmpty checks if there are no error messages.
func (m *Errors) IsEmpty() bool {
	return len(m.messages) == 0
}
