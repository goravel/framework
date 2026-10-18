package mail

import (
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/mail"
)

var _ mail.Fake = &Fake{}

// Fake records the mails sent or queued while faking. One instance is
// shared by every Application created by the service provider, since the
// Mail facade resolves a new Application on each call.
type Fake struct {
	mu      sync.RWMutex
	enabled bool
	sent    []mail.Message
	queued  []mail.Message
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertNotQueued(assertion func(message mail.Message) bool) bool {
	return len(r.Queued(assertion)) == 0
}

func (r *Fake) AssertNotSent(assertion func(message mail.Message) bool) bool {
	return len(r.Sent(assertion)) == 0
}

func (r *Fake) AssertNothingOutgoing() bool {
	return r.AssertNothingSent() && r.AssertNothingQueued()
}

func (r *Fake) AssertNothingQueued() bool {
	return len(r.Queued()) == 0
}

func (r *Fake) AssertNothingSent() bool {
	return len(r.Sent()) == 0
}

func (r *Fake) AssertQueued(assertion func(message mail.Message) bool) bool {
	return len(r.Queued(assertion)) > 0
}

func (r *Fake) AssertQueuedCount(count int) bool {
	return len(r.Queued()) == count
}

func (r *Fake) AssertSent(assertion func(message mail.Message) bool) bool {
	return len(r.Sent(assertion)) > 0
}

func (r *Fake) AssertSentCount(count int) bool {
	return len(r.Sent()) == count
}

func (r *Fake) Queued(filter ...func(message mail.Message) bool) []mail.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return filterMessages(r.queued, filter)
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = false
	r.sent = nil
	r.queued = nil
}

func (r *Fake) Sent(filter ...func(message mail.Message) bool) []mail.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return filterMessages(r.sent, filter)
}

// enable starts faking with empty records.
func (r *Fake) enable() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = true
	r.sent = nil
	r.queued = nil
}

func (r *Fake) isEnabled() bool {
	if r == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.enabled
}

func (r *Fake) recordSent(message mail.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sent = append(r.sent, message)
}

func (r *Fake) recordQueued(message mail.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queued = append(r.queued, message)
}

func filterMessages(messages []mail.Message, filter []func(message mail.Message) bool) []mail.Message {
	result := make([]mail.Message, 0, len(messages))
	for _, message := range messages {
		if len(filter) == 0 || filter[0] == nil || filter[0](message) {
			result = append(result, message)
		}
	}

	return slices.Clip(result)
}
